package model_error

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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

func Logger(c *pgx.Conn, content string) {
	_, logErr := c.Exec(context.Background(), `INSERT INTO sys_log (content, at) VALUES (@content, @at)`, pgx.NamedArgs{"content": content, "at": time.Now()})

	if logErr != nil {
		fmt.Printf("\nFailed to log content:\n%v\n", content)
		fmt.Printf("\nFailure error:\n%v\n", logErr.Error())
	}
}
