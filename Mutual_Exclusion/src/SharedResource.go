package main

import (
	"fmt"
    "net"
    "time"
	"encoding/json"
	"os"
)

type msg_CS struct {
	Id   int 
	Time int    
	Tipo int
	Msg string
}
func CheckError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
		os.Exit(0)
	}
}
func PrintError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}
func main() {
	Address, err := net.ResolveUDPAddr("udp", ":10001")
	CheckError(err)
	Connection, err := net.ListenUDP("udp", Address)
	CheckError(err)
	defer Connection.Close()
	for {
	//Loop infinito para receber mensagem e escrever todo
	//conteúdo (processo que enviou, relógio recebido e texto)
	//na tela
	//FALTA FAZER
		buf := make([]byte, 1024)
		n, _, err := Connection.ReadFromUDP(buf)
			
		if err != nil {
			fmt.Println("Error: ",err)
		}

		msg := buf[0:n]
		resp := msg_CS{}
		json.Unmarshal(msg, &resp)
		fmt.Printf("Processo id = %d entrou na CS. logicalClock = %d. Mensagem enviada: %s\n", resp.Id, resp.Time, resp.Msg)
		time.Sleep(time.Second * 1)
		
	}
}