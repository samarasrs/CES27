package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"math"
	"encoding/json"
	"sync"
)

type msg_item struct {
	Id   int 
	Time int    
	Tipo int
	Msg string
}

//Variáveis globais interessantes para o processo
var err string
var myPort string //porta do meu servidor
var nServers int //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores
 //dos outros processos
var ServConn *net.UDPConn //conexão do meu servidor (onde recebo
 //mensagens dos outros processos)

var CSConn *net.UDPConn

var id int // id dos processos
var logicalClock int
var state int 
var time_request int
var queue []msg_item
var n_reply int
var m sync.Mutex

const (
	RELEASED = 0
	WANTED = 1
	HELD = 2
	REQUEST = 3
	REPLY = 4
	CS = 5
)


func sendMsg(item msg_item, otherProcess int){
	
	msg, _ := json.Marshal(item)
	
	buf := []byte(msg)
	_,err := CliConn[otherProcess - 1].Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}
}

func WantCS(){
	m.Lock()
	logicalClock++
	m.Unlock()
	fmt.Printf("Id = %d SEND REQUEST. Updated logicalClock  = %d.\n", id, logicalClock)
	state = WANTED
	time_request = logicalClock
	item := msg_item{
		Id: id,
		Time: time_request,
		Tipo: REQUEST,
		Msg: "",
	}
	
	for i := 1; i <= nServers; i++ {
		if i != id {
			go sendMsg(item, i)
		}
	}

}
func exitCS(){
	fmt.Printf("Id = %d Sai da CS. logicalClock = %d\n" , id, logicalClock)
	state = RELEASED
	time_request = 0
	for (len(queue)) != 0 {
		dest := queue[0]
		queue = queue[1:]

		item := msg_item{
			Id: id,
			Time: logicalClock,
			Tipo: REPLY,
			Msg: "",

		}
		fmt.Printf("Id = %d SEND REPLY to Process =  %d. logicalClock  = %d.\n", id, dest.Id, logicalClock) 
		go sendMsg(item, dest.Id)
	}
}
func recMsg(resp msg_item){
	m.Lock()
	logicalClock_rec := resp.Time
	logicalClock = int(math.Max(float64(logicalClock_rec), float64(logicalClock))) + 1
	m.Unlock()
	
	if resp.Tipo == REPLY{
		n_reply++
		fmt.Printf("Id = %d RECEIVED REPLY from Process =  %d. Number of replys received = %d . Updated logicalClock = %d.\n", id , resp.Id, n_reply, logicalClock)
		if n_reply == (nServers - 1) {
			n_reply = 0
			state = HELD
			item := msg_item{
				Id: id,
				Time: logicalClock,
				Tipo: CS,
				Msg: "Olá =)",
			}
			msg, _ := json.Marshal(item)
			buf := []byte(msg)
			fmt.Printf("Id = %d Entrei na CS. logicalClock = %d\n", id, logicalClock)
			_,err := CSConn.Write(buf)
			if err != nil {
				fmt.Println(msg, err)
			}
			
			time.Sleep(time.Second * 5)
			go exitCS()
		}

	} else if resp.Tipo == REQUEST {
		fmt.Printf("Id = %d RECEIVED REQUEST from (T = %d, Id = %d). My state = %d and My T = %d. Updated logicalClock = %d. \n", id, resp.Time, resp.Id, state, time_request, logicalClock)
		aux := false
		if time_request < resp.Time{
			aux = true
		} else if time_request == resp.Time {
			if id > resp.Id{
				aux = true
			}
		}

		if (state == HELD || (state == WANTED && aux)){
			queue = append(queue, resp)
		} else{
			item := msg_item{
				Id: id,
				Time: logicalClock,
				Tipo: REPLY,
				Msg: "",
			}
			fmt.Printf("Id = %d SEND REPLY to Process =  %d. logicalClock  = %d.\n", id, resp.Id, logicalClock) 
			go sendMsg(item, resp.Id)
		}
	}
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
func doServerJob() {

	for {
		//Ler (uma vez somente) da conexão UDP a mensagem
		buf := make([]byte, 2048)
		n, _, err := ServConn.ReadFromUDP(buf)
		
		if err != nil {
            fmt.Println("Error: ",err)
        }

		var resp msg_item
		err = json.Unmarshal(buf[0:n], &resp)
		if err != nil {
            fmt.Println("Error: ",err)
        }
		go recMsg(resp)	       
	}
}
func doClientJob(msg string) {

	str_id := strconv.Itoa(id)
	if (state == HELD || state == WANTED){
		fmt.Println("x ignorado.")		
	} else if  msg == str_id {
		m.Lock()
		logicalClock++
		m.Unlock()
		fmt.Printf("Id = %d INTERNAL EVENT. Updated logicalClock atualizado: %d\n", id, logicalClock)
	} else{
		go WantCS()
	}
}

func initConnections() {
	state = RELEASED
	logicalClock = 0
	time_request = 0
	n_reply = 0
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

	CSServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1:10001")
    CheckError(err)

    Conn, err := net.DialUDP("udp", nil, CSServerAddr)
	CSConn = Conn
    CheckError(err)
}
	
func main() {
	
	initConnections()
	
	//O fechamento de conexões deve ficar aqui, assim só fecha
	//conexão quando a main morrer
	defer ServConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}
	defer CSConn.Close()

	ch := make(chan string) //canal que guarda itens lidos do teclado
	go readInput(ch) //chamar rotina que ”escuta” o teclado
	go doServerJob()
	for {
		// Verificar (de forma não bloqueante) se tem algo no
		// stdin (input do terminal)
		select {
			case x, valid := <-ch:
			if valid {
				fmt.Printf("Recebi do teclado: %s \n\n", x)
				go doClientJob(x)
				
			} else {
				fmt.Println("Canal fechado!")
			}
			default:
				// Fazer nada...
				// Mas não fica bloqueado esperando o teclado
				time.Sleep(time.Second * 1)
		}
		// Esperar um pouco
		//time.Sleep(time.Second * 1)
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