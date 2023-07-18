package flix

import (
	"context"
	"fmt"
	"strings"

	"github.com/onflow/flixkit-go"
	"github.com/onflow/flow-cli/flowkit"
)

func GetFlix(args []string, action string, readerWriter flowkit.ReaderWriter) (*flixkit.FlowInteractionTemplate, []string, error) {
	commandParts := strings.Split(action, ":")

	if len(commandParts) != 3 {
		return nil, nil, fmt.Errorf("invalid flix command")
	}

	flixFindMethod := commandParts[1]
	flixIdentifier := commandParts[2]

	var flixService = flixkit.NewFlixService(&flixkit.Config{})
	var template *flixkit.FlowInteractionTemplate
	var argsArr []string

	switch flixFindMethod {
	case "name":
		argsArr = args[1:]
		ctx := context.Background()
		flixRes, err := flixService.GetFlix(ctx, flixIdentifier)
		if err != nil {
			return nil, nil, fmt.Errorf("could not find flix template")
		}
		template = flixRes

	case "id":
		argsArr = args[1:]
		ctx := context.Background()
		flixRes, err := flixService.GetFlixByID(ctx, flixIdentifier)
		if err != nil {
			return nil, nil, fmt.Errorf("could not find flix template")
		}
		template = flixRes

	case "local":
		if flixIdentifier == "path" {
			filePath := args[1]
			argsArr = args[2:]

			flixFile, err := readerWriter.ReadFile(filePath)
			if err != nil {
				return nil, nil, fmt.Errorf("error loading script file: %w", err)
			}

			flixRes, err := flixkit.ParseFlix(string(flixFile))
			if err != nil {
				return nil, nil, fmt.Errorf("error parsing script file: %w", err)
			}

			template = flixRes
		} else {
			return nil, nil, fmt.Errorf("invalid flix command")
		}

	default:
		return nil, nil, fmt.Errorf("invalid flix command")
	}

	return template, argsArr, nil
}
