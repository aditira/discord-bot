package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func main() {
	godotenv.Load(".env")

	client := openai.NewClient(os.Getenv("OPENAI_TOKEN"))
	messages := []openai.ChatCompletionMessage{}

	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_KEY"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		fmt.Println(m.Author.Username + " | " + m.Content)
		if m.Author.ID != s.State.User.ID && m.Content != "" {

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: m.Content,
			})

			resp, err := client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model:    openai.GPT3Dot5Turbo,
					Messages: messages,
				},
			)

			if err != nil {
				panic(err)
			}

			content := resp.Choices[0].Message.Content
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: content,
			})

			if ch, err := s.State.Channel(m.ChannelID); err != nil || !ch.IsThread() {
				_, _ = s.ChannelMessageSend(m.ChannelID, content)
			} else {
				_, _ = s.ChannelMessageSendReply(m.ChannelID, content, m.Reference())
			}
		}

	})

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()

}

// https://github.com/bwmarrin/discordgo
// https://github.com/sashabaranov/go-openai
