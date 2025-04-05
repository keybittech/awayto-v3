package clients

import (
	"context"
	"errors"
	"time"
)

type CommandHandler[Command any, Params any, Response any] interface {
	GetCommandChannel() chan<- Command
}

func SendCommand[Command any, Params any, Response any](
	handler CommandHandler[Command, Params, Response],
	createCommand func(Params, chan Response) Command,
	params Params,
) (Response, error) {
	var emptyResponse Response

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	replyChan := make(chan Response)

	cmd := createCommand(params, replyChan)

	select {
	case handler.GetCommandChannel() <- cmd:
	case <-ctx.Done():
		return emptyResponse, errors.New("timed out when sending command")
	}

	select {
	case response := <-replyChan:
		return response, nil
	case <-ctx.Done():
		return emptyResponse, errors.New("timed out when receiving command")
	}
}

func ChannelError(general, response error) error {
	if general != nil {
		return general
	} else if response != nil {
		return response
	}

	return nil
}
