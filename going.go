package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/subcommands"
	"gopkg.in/yaml.v2"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&initCmd{}, "")
	subcommands.Register(&upCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}

///// init /////
type initCmd struct {
}

func (*initCmd) Name() string     { return "init" }
func (*initCmd) Synopsis() string { return "init" }
func (*initCmd) Usage() string {
	return `init:
  it makes files to use other commands.
`
}

func (p *initCmd) SetFlags(f *flag.FlagSet) {
}

func (p *initCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// TODO: ここのディレクトリ変更
	path := os.Getenv("GOPATH") + "/src/" + "github.com/islands5/going/assets"
	err := exec.Command("cp", "-r", path, "./going-assets").Run()
	if err != nil {
		fmt.Println(err)
	}
	return subcommands.ExitSuccess
}

///// up /////
type upCmd struct {
}

func (*upCmd) Name() string     { return "up" }
func (*upCmd) Synopsis() string { return "execute db migration" }
func (*upCmd) Usage() string {
	return `up:
  it executes db migration and records finished *.sql.
`
}

func (p *upCmd) SetFlags(f *flag.FlagSet) {
}

func (p *upCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	connInfo, err := loadYml("going-assets/going.yml")

	db, err := connectMysql(connInfo["db_name"], connInfo["user"], connInfo["password"])
	defer db.Close()

	files, err := ioutil.ReadDir("sql")

	_, err = db.Query("CREATE DATABASE test_p")
	if err != nil {
		fmt.Println(err)
	}

	return subcommands.ExitSuccess
}

///// reset /////
type resetCmd struct {
}

func connectMysql(database, user, password interface{}) (*sql.DB, error) {
	conn := fmt.Sprintf("%s:%s@/%s", user, password, database)

	db, err := sql.Open("mysql", conn)
	if err != nil {
		fmt.Println(err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println(err)
	}

	return db, err
}

func loadYml(filename string) (map[interface{}]interface{}, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}

	connInfo := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &connInfo)
	if err != nil {
		fmt.Println(err)
	}
	return connInfo, err
}
