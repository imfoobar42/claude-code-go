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
	resp, err := client.Chat.Completions.New(context.Background(),
		openai.ChatCompletionNewParams{
			Model: "anthropic/claude-haiku-4.5",
			Messages: []openai.ChatCompletionMessageParamUnion{
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.String(prompt),
						},
					},
				},
			},
			Tools: []openai.ChatCompletionToolUnionParam{
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
			},
		},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Choices) == 0 {
		panic("No choices in response")
	}
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		toolCall := resp.Choices[0].Message.ToolCalls[0]

		if toolCall.Function.Name != "Read" || toolCall.Type != "function" {
			panic("Unexpected tool call")
		}

		//Parse arguments
		var arguments map[string]any
		if err := json.Unmarshal(
			[]byte(toolCall.Function.Arguments), &arguments); err != nil {
			panic(fmt.Sprintf("Failed to parse tool call arguments: %v", err))
		}

		var filePath string
		if v, ok := arguments["file_path"].(string); ok {
			filePath = v
		} else {
			panic("file_path argument missing or not a string")
		}
		//Read file contents
		content, err := os.ReadFile(filePath)
		if err != nil {
			panic(fmt.Sprintf("Failed to read file: %v", err))
		}
		fmt.Printf("%s\n", string(content))
	}

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	// TODO: Uncomment the line below to pass the first stage
	fmt.Print(resp.Choices[0].Message.Content)
}
