package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
)

// input prompts the user for input and returns the result
func input(message string) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
	}
	err := survey.AskOne(prompt, &result)
	if err != nil {
		if err.Error() == "interrupt" {
			return "", errors.New("cancelled by user")
		}
		return "", err
	}
	return result, nil
}

// inputWithDefault prompts for input with a default value
func inputWithDefault(message, defaultValue string) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}
	err := survey.AskOne(prompt, &result)
	if err != nil {
		if err.Error() == "interrupt" {
			return "", errors.New("cancelled by user")
		}
		return "", err
	}
	return result, nil
}

// passwordInput prompts for password/token input (hidden)
func passwordInput(message string) (string, error) {
	var result string
	prompt := &survey.Password{
		Message: message,
	}
	err := survey.AskOne(prompt, &result)
	if err != nil {
		if err.Error() == "interrupt" {
			return "", errors.New("cancelled by user")
		}
		return "", err
	}
	return result, nil
}

// selectOption shows a list of options and returns the selected index
func selectOption(message string, options []string) (int, error) {
	var result string
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}

	// Customize the prompt to remove the "?" icon
	surveyIcon := &survey.IconSet{}
	surveyIcon.Question = survey.Icon{Text: ""}

	err := survey.AskOne(prompt, &result, survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question = survey.Icon{Text: ""}
	}))
	if err != nil {
		if err.Error() == "interrupt" {
			return -1, errors.New("cancelled by user")
		}
		return -1, err
	}

	// Find the index of the selected option
	for i, option := range options {
		if option == result {
			return i, nil
		}
	}
	return -1, fmt.Errorf("selected option not found")
}

// selectMultiOption shows a list of options and returns selected indices
func selectMultiOption(message string, options []string) ([]int, error) {
	var results []string
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
	}
	err := survey.AskOne(prompt, &results)
	if err != nil {
		if err.Error() == "interrupt" {
			return nil, errors.New("cancelled by user")
		}
		return nil, err
	}

	// Find indices of selected options
	var indices []int
	for _, result := range results {
		for i, option := range options {
			if option == result {
				indices = append(indices, i)
				break
			}
		}
	}
	return indices, nil
}

// confirm asks for yes/no confirmation
func confirm(message string) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
		Default: false,
	}
	err := survey.AskOne(prompt, &result)
	if err != nil {
		if err.Error() == "interrupt" {
			return false, errors.New("cancelled by user")
		}
		return false, err
	}
	return result, nil
}

// askNumber asks for a number input
func askNumber(message string) (int, error) {
	result, err := input(message)
	if err != nil {
		return 0, err
	}

	num, err := strconv.Atoi(result)
	if err != nil {
		return 0, fmt.Errorf("please enter a valid number")
	}
	return num, nil
}

// askFloat asks for a float input
func askFloat(message string) (float64, error) {
	result, err := input(message)
	if err != nil {
		return 0, err
	}

	num, err := strconv.ParseFloat(result, 64)
	if err != nil {
		return 0, fmt.Errorf("please enter a valid number")
	}
	return num, nil
}
