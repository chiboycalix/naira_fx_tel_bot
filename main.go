package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/adrg/exrates"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

// Send any text message to the bot after the bot has been started

func BotToken() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return os.Getenv("BOT_TOKEN")
}

type Envelope struct {
	Cube []struct {
		Date  string `xml:"time,attr"`
		Rates []struct {
			Currency string `xml:"currency,attr"`
			Rate     string `xml:"rate,attr"`
		} `xml:"Cube"`
	} `xml:"Cube>Cube"`
}

func main() {
	resp, err := http.Get("http://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml")

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	xmlCurrenciesData, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var env Envelope
	err = xml.Unmarshal(xmlCurrenciesData, &env)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Date ", env.Cube[0].Date)

	for _, v := range env.Cube[0].Rates {
		fmt.Println("Currency : ", v.Currency, " Rate : ", v.Rate)
	}

	// REMEMBER! this data is from European Central Bank
	// therefore the rates are based on EUR

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(BotToken(), opts...)
	if nil != err {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/hello", bot.MatchTypeExact, helloHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/print_date", bot.MatchTypeExact, printDateHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/get_naira_rate", bot.MatchTypeExact, printNairaRates)

	b.Start(ctx)
}

func helloHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Hello, *" + bot.EscapeMarkdown(update.Message.From.FirstName) + "*",
		ParseMode: models.ParseModeMarkdown,
	})
}
func printDateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   time.Now().String(),
	})
}
func printNairaRates(ctx context.Context, b *bot.Bot, update *models.Update) {
	rates, err := exrates.Latest("USD", nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Exchange rates for %s on %s\n", rates.Base, rates.Date)
	for currency, value := range rates.Values {
		fmt.Printf("%s: %f\n", currency, value)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   currency + strconv.FormatFloat(value, 'f', 2, 64),
		})
	}
}
func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Say /hello",
	})
}
