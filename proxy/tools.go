package proxy

import (
	"context"
	"io"
	"net"
	"os"
)

func JoinConnection(ctx context.Context, src, dest net.Conn, errChan chan error)error{
	_, err := io.Copy(src, dest)
	return SendError(errChan, err)
}

func SendError(errChan chan error, err error)error{
	if err != nil {
		go func(){errChan <- err}()
	}
	return err
}

//检查文件是否存在，不存在就创建
func KeepFileExists(name string)error{
	file, err := os.OpenFile(name, os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func SendConn(connChan chan Connection, conn Connection){
	go func(){connChan <- conn}()
}