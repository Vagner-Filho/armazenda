package model_error

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ModelError struct {
	IsServerErr bool
	Message     string
}

func (e *ModelError) Error() string {
	return e.Message
}

func (e *ModelError) FromServer() bool {
	return e.IsServerErr
}

type loggerModel struct {
	pool *pgxpool.Pool
}

var loggerModelImpl *loggerModel

func InitLoggerModel(pool *pgxpool.Pool) (*loggerModel, error) {
	if pool == nil {
		return nil, errors.New("pool cant be null")
	}

	if loggerModelImpl == nil {
		loggerModelImpl = &loggerModel{
			pool: pool,
		}
	}

	return loggerModelImpl, nil
}

func GetLoggerModel() *loggerModel {
	if loggerModelImpl == nil {
		panic("\nlogger model hasnt been initialized\n")
	}
	return loggerModelImpl
}

func (lm *loggerModel) Log(content string) {
	_, logErr := lm.pool.Exec(context.Background(), `INSERT INTO sys_log (content, at) VALUES (@content, @at)`, pgx.NamedArgs{"content": content, "at": time.Now()})

	if logErr != nil {
		fmt.Printf("\nFailed to log content:\n%v\n", content)
		fmt.Printf("\nFailure error:\n%v\n", logErr.Error())
	}
}
