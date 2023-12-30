# RaftKV

This is a simple kv database writting in go. And it is a student project for Distributed System Course in SYSU.

## Usage

- Use `db := kvnode.Open()` to create a new database `db` and **don't forget** to use `db.Close()` to close the database and persist the data.
  - use `db.Put()`,`db.remove()`,`db.update()` and `db.get()` to manage your data.
- Use `config.yaml` to change the config of service.
  - `mode` can be set to *default* or *config*. When mode is `default`, it will use the config in `kvserver.go`
  - `serversNum` determines how many nodes will be created.
  - `address` and `port` combine to be node's address.
  - `persistPath` detemines where to persist data. 
- Communication between nodes uses grpc. `proto` file is in the ./rpc. If you edit it, pleause use the command below to generate the new gprc file and then move the files into ./raftkv
- You can set `Debug=true` in file debug/util.go, and use `debug.Dlog` to debug.

```shell
protoc --proto_path=./rpc --go_out=./kv --go-grpc_out=./kv ./rpc/kvrpc.proto
```
## Demo
```go

func main() {
	db := kvserver.Open()
	go randOp(db)
	go randOp(db)
	db.Close()
}

func randOp(db *kvserver.DBService) {

	oplis := [4]string{"get", "put", "remove", "update"}
	keylis := [20]string{"a", "b", "c", "se", "aa", "Esx", "ky", "mio", "ss", "teio", "rs", "sada", "mc", "mq", "xxx23x", "asddaw", "xxas", "3a1sd", "liky", "fll"}
	vallis := [20]int{2, 96, -2, 5, 33, -88, 5, 23, 6, 845, 9, 2, 922, 912123, 223, 55, 31, 5223, 5, 2631}
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(rand.Int()%30) * time.Millisecond)
		op := oplis[rand.Int()%4]
		key := keylis[rand.Int()%8]
		val := vallis[rand.Int()%10]
		fmt.Println(op, " ", key, " ", val)
		if op == "get" {
			_, err := db.Get(key)
			if err != nil {
				panic(err)
			}
		} else if op == "put" {
			err := db.Put(key, val)
			if err != nil {
				panic(err)
			}
		} else if op == "update" {
			err := db.Update(key, val)
			if err != nil {
				panic(err)
			}
		} else {
			err := db.Remove(key)
			if err != nil {
				panic(err)
			}
		}
	}
}

```