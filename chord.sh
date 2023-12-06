#!/bin/bash

go build -o build/chord

# Initialize variables
inputA=""
inputJ=""
inputTCP=""
inputTS=""
inputFF=""

# Process arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -a)
            inputA="$2"
            shift # remove flag
            shift # remove value
            ;;
        -j)
            inputJ="$2"
            shift
            shift
            ;;
        -tcp)
            inputTCP="$2"
            shift
            shift
            ;;
        -ts)
            inputTS="$2"
            shift
            shift
            ;;
        -ff)
            inputFF="$2"
            shift
            shift
            ;;
        *)
            echo "Invalid argument: $1"
            exit 1
            ;;
    esac
done

# Check if all required variables are set
if [ -z "$inputA" ] || [ -z "$inputTCP" ] || [ -z "$inputTS" ] || [ -z "$inputFF" ]; then
    echo "Usage: $0 -a inputA [-j inputJ] -tcp inputTCP -ts inputTS -ff inputFF"
    exit 1
fi

# Construct the command
command="build/chord -a \"$inputA\""
[ -n "$inputJ" ] && command+=" -j \"$inputJ\""
command+=" -tcp \"$inputTCP\" -ts \"$inputTS\" -ff \"$inputFF\""

# Run the chord command with the provided arguments
eval $command
