package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"math"
)

//Variáveis globais interessantes para o processo
var err string
var myPort string //porta do meu servidor
var nServers int //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores
 //dos outros processos
var ServConn *net.UDPConn //conexão do meu servidor (onde recebo
 //mensagens dos outros processos)

var id int // id dos processos
var logicalClock int

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
func doServerJob() {
    buf := make([]byte, 1024)
	
	
	

	for {
		//Ler (uma vez somente) da conexão UDP a mensagem
		//Escrever na tela a msg recebida (indicando o
		//endereço de quem enviou)
		n, _, err := ServConn.ReadFromUDP(buf)
		msg := string(buf[0:n])
		logicalClock_rec, _ := strconv.Atoi(msg)
		logicalClock = int(math.Max(float64(logicalClock_rec), float64(logicalClock))) + 1
        fmt.Println("Valor do logicalClock recebido: ", msg)
		fmt.Printf("logicalClock atualizado: %d\n", logicalClock)
        if err != nil {
            fmt.Println("Error: ",err)
        } 
	}
}

func doClientJob(otherProcess int) {
	//Enviar uma mensagem (com valor i) para o servidor do processo
	//otherServer
	msg := strconv.Itoa(logicalClock)
	if otherProcess == id {
		fmt.Println("Internal Event. \nlogicalClock atualizado: " + msg)
	} else{
		
		buf := []byte(msg)
		_,err := CliConn[otherProcess - 1].Write(buf)
		fmt.Println("Valor do logicalClock enviado: " + msg)
		if err != nil {
			fmt.Println(msg, err)
		}
	}
}

func initConnections() {
	logicalClock = 0
	id, _ = strconv.Atoi(os.Args[1])
	myPort = os.Args[id + 1]
	nServers = len(os.Args) - 2
	/*Esse 2 tira o nome (no caso Process) e tira a primeira porta (que
	é a minha). As demais portas são dos outros processos*/
	CliConn = make([]*net.UDPConn, nServers)
	//Outros códigos para deixar ok a conexão do meu servidor (onde recebo msgs). O processo já deve ficar habilitado a receber msgs.
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1"+myPort)
	CheckError(err)
	ServConn, err = net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	//Outros códigos para deixar ok as conexões com os servidores dos
	//outros processos. Colocar tais conexões no vetor CliConn.
	for servidores := 0; servidores < nServers; servidores++ {
		ServerAddr, err := net.ResolveUDPAddr("udp",
		"127.0.0.1"+os.Args[2+servidores])
		CheckError(err)
		Conn, err := net.DialUDP("udp", nil, ServerAddr)
		CliConn[servidores] = Conn
		CheckError(err)
	}
}
	
func main() {
	initConnections()
	//O fechamento de conexões deve ficar aqui, assim só fecha
	//conexão quando a main morrer
	defer ServConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}

	ch := make(chan string) //canal que guarda itens lidos do teclado
	go readInput(ch) //chamar rotina que ”escuta” o teclado
	go doServerJob()
	for {
		// Verificar (de forma não bloqueante) se tem algo no
		// stdin (input do terminal)
		select {
			case x, valid := <-ch:
			if valid {
				fmt.Printf("Recebi do teclado: %s \n", x)
				id_dest, _ := strconv.Atoi(x)
				logicalClock++
				go doClientJob(id_dest)
				
			} else {
				fmt.Println("Canal fechado!")
			}
			default:
				// Fazer nada...
				// Mas não fica bloqueado esperando o teclado
				time.Sleep(time.Second * 1)
		}
		// Esperar um pouco
		time.Sleep(time.Second * 1)
	}

}

func readInput(ch chan string) {
	// Rotina não-bloqueante que “escuta” o stdin
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}