package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

const (
	mapPath string = "mapping.txt"
	artPath string = "ascii-art.txt"
)

func main() {
	mapping, err := loadMap()
	if err != nil {
		log.Fatal(err)
	}

	rows, err := loadRows()
	if err != nil {
		log.Fatal(err)
	}

	plaintext := make([]string, len(rows))
	for src, pos := range mapping {
		if src >= len(plaintext) {
			log.Fatalf("wrong row starting position")
		}
		if pos >= len(mapping) {
			log.Fatalf("wrong row final position")
		}
		plaintext[src] = rows[pos]
	}

	for _, row := range plaintext {
		fmt.Println(row)
	}
}

func loadMap() (map[int]int, error) {
	file, err := os.Open(mapPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rows := make(map[int]int)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var src, pos int
		fmt.Sscanf(scanner.Text(), "%d: %d", &src, &pos)
		rows[pos] = src
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rows, nil
}

func loadRows() ([]string, error) {
	file, err := os.Open(artPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rows := make([]string, 27)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rows = append(rows, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rows, nil
}
