package main

import (
	"context" //used to control cancellations, timeouts and request lifetimes
	"encoding/json"
	"flag" //for parsing command line arguments
	"fmt"  //for formatted I/O
	"os"   //for interacting with the operating system, such as environment variables and standard input/output

	//for encoding and decoding JSON data

	"github.com/openai/openai-go/v3"        //open AI client SDK
	"github.com/openai/openai-go/v3/option" //for configuring the OpenAI key, base URL
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse() //read the command line arguments and populate the prompt variable

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}
	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))
	messages := []openai.ChatCompletionMessageParamUnion{
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(prompt),
				},
			},
		},
	}
	tools := []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Read",
			Description: openai.String("Read and return the contents of the file"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "The path of the file to read, relative to the current working directory.",
					},
				},
				"required": []string{"file_path"},
			},
		}),
	}

	for {
		resp, err := client.Chat.Completions.New(context.Background(),
			openai.ChatCompletionNewParams{
				Model:    "anthropic/claude-haiku-4.5",
				Messages: messages,
				Tools:    tools,
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(resp.Choices) == 0 {
			panic("No choices in response")
		}

		message := resp.Choices[0].Message
		var toolCalls []openai.ChatCompletionMessageToolCallUnionParam
		for _, tc := range message.ToolCalls {
			toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID: tc.ID,
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				},
			})
		}
		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Content:   openai.String(message.Content),
				ToolCalls: toolCalls,
			},
		})

		if len(message.ToolCalls) == 0 {
			fmt.Print(message.Content)
			break
		}

		for _, toolCall := range message.ToolCalls {
			if toolCall.Function.Name != "Read" || toolCall.Type != "function" {
				panic("Unexpected tool call")
			}

			var arguments map[string]any
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &arguments); err != nil {
				panic(fmt.Sprintf("Failed to parse tool call arguments: %v", err))
			}

			filePath, ok := arguments["file_path"].(string)
			if !ok {
				panic("file_path argument missing or not a string")
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				panic(fmt.Sprintf("Failed to read file: %v", err))
			}

			messages = append(messages, openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(string(content)),
					},
					ToolCallID: toolCall.ID,
				},
			})
		}
	}
}
