package main

import (
	"log"

	"github.com/spf13/cobra"
	"proxy.bing89.com/client"
	"proxy.bing89.com/server"
)


func printLog(errChan chan error){
	for {
		select{
		case err := <- errChan:
			log.Println(err)
		}
	}
}

func main(){
	errChan := make(chan error)
	var (
		port *string
		proxyConfigFile *string
	)
	
	serverCommand := cobra.Command{
		Short: "server",
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			svr, err := server.NewServer(*port, *proxyConfigFile, errChan)
			if err != nil {
				log.Fatal(err)
			}
			svr.Run()
		},
	}
	// server = serverCommand.Flags().StringP("server", "s", "", "server ip addrss")
	port = serverCommand.Flags().StringP("port", "p", "", "server port")
	proxyConfigFile = serverCommand.Flags().StringP("config", "c", "proxy.json", "proxy config file")
	var (
		serverIP *string
		serverPort *string
		name *string
	)
	
	clientCommand := cobra.Command{
		Short: "client",
		Use: "client",
		Run: func(cmd *cobra.Command, args []string) {
			clt, err := client.NewClient(*name, *serverIP, *serverPort)
			if err != nil {
				log.Fatal(err)
			}
			err = clt.Run(errChan)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	serverIP = clientCommand.Flags().StringP("server", "s", "", "server ip addrss")
	serverPort = clientCommand.Flags().StringP("port", "p", "", "server port")
	name = clientCommand.Flags().StringP("name", "n", "", "client name")

	cmd := cobra.Command{
		Short: "proxy [server] [client] [flags]",
		Use: "proxy",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("nothing to do")
		},
	}
	cmd.AddCommand(&serverCommand, &clientCommand)
	go printLog(errChan)
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}