Adicione a pasta labMapReduce no seu $GOPATH/src

o $GOPATH é o diretório onde você armazena seus códigos Go.

# Para compilar:
Entrar em ```$GOPATH/src/labMapReduce/wordcount```
Executar: ```go build```

# Para rodar:
Entrar em ```$GOPATH/src/labMapReduce/wordcount```

## Sequetial:
Executar: ```wordcount.exe -mode sequential -file files/teste.txt -chunksize 100 -reducejobs 2```

Obs1: Esse arquivo teste.txt é o que será processado e está na pasta files.

Obs2: Os valores de ```-chunksize``` e ```-reducejobs``` podem ser modificados conforme o desejado, bem como o arquivo de entrada.

## Distributed:
### workers:
```wordcount.exe -mode distributed -type worker –port 50001```

Obs1: para executar mais de um worker basta executar em outro terminal o comando acima com ```–port 50002```.

Obs2: para forçar o worker a falhar basta acrescentar ao comando acima: ```-fail x```.

Onde x corresponde ao valor (inteiro) da operção que o worker irá falhar (Se x = 3 o worker irá falaha na sua terceira tarefa).

### master:
```wordcount.exe -mode distributed -type master -file files/pg1342.txt -chunksize 102400 -reducejobs 5```

