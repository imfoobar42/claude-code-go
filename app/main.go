package main

import (
	"context" //used to control cancellations, timeouts and request lifetimes
	"encoding/json"
	"flag"    //for parsing command line arguments
	"fmt"     //for formatted I/O
	"os"      //for interacting with the operating system, such as environment variables and standard input/output
	"os/exec" //for running shell commands
	"strings" //for splitting command arguments

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
	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl)) //initialize OpenAI client
	messages := []openai.ChatCompletionMessageParamUnion{                              //conversation history that grows each loop
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(prompt),
				},
			},
		},
	}
	tools := []openai.ChatCompletionToolUnionParam{ //tool definitions exposed to the model
		//Read tool definition
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
		//Write tool definition
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Write",
			Description: openai.String("Write content to a file"),
			Parameters: openai.FunctionParameters{
				"type":     "object",
				"required": []string{"file_path", "content"},
				"properties": map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "Path of the file to write to",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "Content to write to the file",
					},
				},
			},
		}),
		//Bash tool definition
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Bash",
			Description: openai.String("Execute shell commands"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "Command to execute",
					},
				},
				"required": []string{"command"},
			},
		}),
	}

	for { //agent loop: keep calling the model until no tool calls remain
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

		message := resp.Choices[0].Message //assistant response for this turn
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
		messages = append(messages, openai.ChatCompletionMessageParamUnion{ //store assistant message in history
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: openai.String(message.Content),
				},
				ToolCalls: toolCalls,
			},
		})

		if len(message.ToolCalls) == 0 { //no tool calls means the model is done
			fmt.Print(message.Content)
			break
		}

		for _, toolCall := range message.ToolCalls { //execute each requested tool call
			if toolCall.Type != "function" {
				panic("Unexpected tool call type")
			}

			var arguments map[string]any
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &arguments); err != nil {
				panic(fmt.Sprintf("Failed to parse tool call arguments: %v", err))
			}

			toolOutput := ""
			switch toolCall.Function.Name {
			case "Read": //Read: return file contents
				filePath, ok := arguments["file_path"].(string)
				if !ok {
					panic("file_path argument missing or not a string")
				}
				content, err := os.ReadFile(filePath)
				if err != nil {
					panic(fmt.Sprintf("Failed to read file: %v", err))
				}
				toolOutput = string(content)
			case "Write": //Write: overwrite or create file
				filePath, ok := arguments["file_path"].(string)
				if !ok {
					panic("file_path argument missing or not a string")
				}
				content, ok := arguments["content"].(string)
				if !ok {
					panic("content argument missing or not a string")
				}
				if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
					panic(fmt.Sprintf("Failed to write file: %v", err))
				}
				toolOutput = "OK"
			case "Bash": //Bash: execute shell command and return output
				command, ok := arguments["command"].(string)
				if !ok {
					panic("command argument missing or not a string")
				}
				parts := strings.Fields(command)
				if len(parts) == 0 {
					panic("command argument missing or empty")
				}
				outputBytes, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
				toolOutput = string(outputBytes)
				if err != nil {
					toolOutput += fmt.Sprintf("\nCommand error: %v", err)
				}
			default:
				panic("Unexpected tool call")
			}

			messages = append(messages, openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(toolOutput),
					},
					ToolCallID: toolCall.ID,
				},
			})
		}
	}
}
