package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/starslipay/account_mgr/internal/consts"
	"github.com/starslipay/account_mgr/internal/svc"
	"github.com/starslipay/account_mgr/model/mysql"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type C2CTransferMessage struct {
	TransactionId string `json:"transaction_id"`
	BuyerUid      int64  `json:"buyer_uid"`
	BuyerUserId   string `json:"buyer_user_id"`
	SellerUid     int64  `json:"seller_uid"`
	SellerUserId  string `json:"seller_user_id"`
	Amount        int64  `json:"amount"`
	CurType       int32  `json:"cur_type"`
	Desc          string `json:"desc"`
}

type C2CConsumer struct {
	svcCtx        *svc.ServiceContext
	logger        logx.Logger
	wg            sync.WaitGroup
	isRunning     bool
	consumerGroup sarama.ConsumerGroup
}

func NewC2CConsumer(svcCtx *svc.ServiceContext) *C2CConsumer {
	return &C2CConsumer{
		svcCtx: svcCtx,
		logger: logx.WithContext(context.Background()),
	}
}

func (c *C2CConsumer) Start(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Version = sarama.V3_8_0_0
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(c.svcCtx.Config.Kafka.BrokerAddrs, "account_mgr_c2c_consumer", config)
	if err != nil {
		return fmt.Errorf("failed to create consumer group: %v", err)
	}

	c.consumerGroup = consumerGroup
	c.isRunning = true

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for c.isRunning {
			if err := consumerGroup.Consume(ctx, []string{c.svcCtx.Config.Kafka.Topic}, c); err != nil {
				c.logger.Errorf("consumer group error: %v", err)
				time.Sleep(time.Second * 5)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	c.logger.Info("C2C consumer started successfully")
	return nil
}

func (c *C2CConsumer) Stop() {
	c.isRunning = false
	if c.consumerGroup != nil {
		c.consumerGroup.Close()
	}
	c.wg.Wait()
	c.logger.Info("C2C consumer stopped")
}

func (c *C2CConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *C2CConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *C2CConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		c.logger.Infof("Received message: topic=%s, partition=%d, offset=%d, key=%s",
			message.Topic, message.Partition, message.Offset, string(message.Key))

		err := c.processMessage(string(message.Value))
		if err != nil {
			c.logger.Errorf("Process message failed: %v", err)
			continue
		}

		session.MarkMessage(message, "")
	}

	return nil
}

func (c *C2CConsumer) processMessage(body string) error {
	var msg C2CTransferMessage
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		return fmt.Errorf("unmarshal message failed: %v", err)
	}

	return c.processC2CTransfer(&msg)
}

func (c *C2CConsumer) processC2CTransfer(msg *C2CTransferMessage) error {
	_, err := c.svcCtx.TLocalMessageModelSlave.FindOneByTransactionId(context.Background(), msg.TransactionId)
	if err == nil {
		msgRecord, _ := c.svcCtx.TLocalMessageModelSlave.FindOneByTransactionId(context.Background(), msg.TransactionId)
		if msgRecord.State == consts.MsgStateDone {
			c.logger.Infof("Message already processed: %s", msg.TransactionId)
			return nil
		}
	}

	err = c.svcCtx.SqlMasterConn.TransactCtx(context.Background(), func(ctx context.Context, session sqlx.Session) error {
		tcAccountModel := mysql.NewTCAccountModel(sqlx.NewSqlConnFromSession(session))
		tcAccountLogModel := mysql.NewTCAccountLogModel(sqlx.NewSqlConnFromSession(session))
		tLocalMessageModel := mysql.NewTLocalMessageModel(sqlx.NewSqlConnFromSession(session))

		buyerAccount, err := tcAccountModel.FindOne(ctx, msg.BuyerUid)
		if err != nil {
			return err
		}

		err = tcAccountModel.AddBalance(ctx, msg.BuyerUid, msg.Amount)
		if err != nil {
			return err
		}

		_, err = tcAccountLogModel.Insert(ctx, &mysql.TCAccountLog{
			Uid:                msg.BuyerUid,
			UserId:             msg.BuyerUserId,
			CounterpartyUserId: msg.SellerUserId,
			CounterpartyUid:    msg.SellerUid,
			TransactionId:      msg.TransactionId,
			InoutType:          consts.InoutTypeIn,
			BizType:            consts.BizTypeC2C,
			Amount:             msg.Amount,
			Balance:            buyerAccount.Balance + msg.Amount,
			Desc:               msg.Desc,
		})
		if err != nil {
			return err
		}

		msgRecord, err := tLocalMessageModel.FindOneByTransactionId(ctx, msg.TransactionId)
		if err == nil {
			msgRecord.State = consts.MsgStateDone
			err = tLocalMessageModel.Update(ctx, msgRecord)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.logger.Errorf("Process C2C transfer failed: %v", err)
		return err
	}

	c.logger.Infof("Process C2C transfer success: %s", msg.TransactionId)
	return nil
}
