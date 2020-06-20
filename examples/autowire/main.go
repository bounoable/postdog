package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/bounoable/postdog/autowire"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/template"

	// Register packages for autowiring
	_ "github.com/bounoable/postdog/plugin/markdown/goldmark"
	_ "github.com/bounoable/postdog/plugin/store/memorystore"
	_ "github.com/bounoable/postdog/plugin/template"
	_ "github.com/bounoable/postdog/transport/nop"
	_ "github.com/bounoable/postdog/transport/smtp"
)

func main() {
	wd, _ := os.Getwd()
	os.Setenv("APP_ROOT", wd)

	configPath := filepath.Join(wd, "configs/postdog.yml")
	cfg, err := autowire.File(configPath)
	if err != nil {
		log.Fatal(err)
	}

	po, err := cfg.Office(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Use the "welcome" template for this context.
	ctx := template.Enable(context.Background(), "welcome", nil)

	err = po.Send(ctx, letter.Write(
		letter.Subject("Hello"),
		letter.From("the", "other@side.com"),
		letter.Text("# Hello"),
		letter.HTML("This should be overriden."),
	))

	if err != nil {
		log.Fatal(err)
	}
}
