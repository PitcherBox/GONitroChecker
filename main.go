package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	math "math/rand"

	"github.com/fatih/color"
	"github.com/pelletier/go-toml"
)

var threads = 0
var useProxies = true
var proxyType = "http"
var version = "v1.0"

var invalidPromos = []string{}
var validPromos = []string{}

func main() {

	tokens := readFile("./input/promos.txt")
	proxies := readFile("./input/proxies.txt")

	config, err := toml.LoadFile("./config.toml")
	if err != nil {
		fmt.Println(err)
		return
	}

	threads = int(config.Get("threads").(int64))

	var i string

	fmt.Println("\nPress ENTER to start checking.")
	fmt.Scanln(&i)

	threadingManager(tokens, proxies)

}

func threadingManager(tokens []string, proxies []string) {

	var threadManager sync.WaitGroup
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	var threadWorks = make(map[int][]string)

	if len(tokens) < threads {
		threads = len(tokens)
		fmt.Fprintln(color.Output, green("[", currentTime(), "] Number of threads greater than number of promos, setted in ", fmt.Sprint(threads)))
	}

	fmt.Fprintln(color.Output, green("[", currentTime(), "] Setting task to threads..."))

	var counter = 0

	for i := 0; i < len(tokens); i++ {

		if counter >= threads {
			counter = 0
		}

		threadTokens := threadWorks[counter]
		token := strings.Split(tokens[i], "/")
		threadTokens = append(threadTokens, token[len(token)-1])
		threadWorks[counter] = threadTokens
		counter++

	}

	for i := 0; i < len(threadWorks); i++ {
		fmt.Fprintln(color.Output, cyan("[", currentTime(), "] Thread #"+fmt.Sprint(i+1)+": Created with "+fmt.Sprint(len(threadWorks[i]))+" promos!"))
	}

	fmt.Fprintln(color.Output, green("[", currentTime(), "] Tasks established to threads!"))

	threadManager.Add(threads)

	for i := 0; i < threads; i++ {
		go func(i int) {
			defer threadManager.Done()
			fmt.Fprintln(color.Output, cyan("[", currentTime(), "] Thread #"+fmt.Sprint(i+1)+": Started"))
			thread(i, threadWorks[i], proxies)
		}(i)
	}

	threadManager.Wait()
	fmt.Fprintln(color.Output, green("---------------------------------------------------------------------------"))
	fmt.Fprintln(color.Output, green("Total valid promos: "), cyan(len(validPromos)))
	fmt.Fprintln(color.Output, green("Total invalid promos: "), red(len(invalidPromos)))
	fmt.Fprintln(color.Output, green("---------------------------------------------------------------------------"))

	createFile("./output/valid.txt", validPromos)
}

func thread(id int, tokens []string, proxies []string) {
	var thread sync.WaitGroup
	cyan := color.New(color.FgCyan).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	thread.Add(len(tokens))

	for i := 0; i < len(tokens); i++ {
		go func(i int) {
			defer thread.Done()
			fmt.Fprintln(color.Output, cyan("[", currentTime(), "] Thread #"+fmt.Sprint(id+1)+": Checking promo "+tokens[i]))

			client := &http.Client{}

			if useProxies && len(proxies) > 0 {
				proxyTemplate := proxies[math.Intn(len(proxies))]
				if !strings.Contains(proxyTemplate, proxyType) {
					proxyTemplate = proxyType + "://" + proxyTemplate
				}
				proxyURL, err := url.Parse(proxyTemplate)
				if err == nil {
					transport := &http.Transport{
						Proxy: http.ProxyURL(proxyURL),
					}
					client = &http.Client{
						Transport: transport,
					}
				}
			}

			req, err := http.NewRequest("GET", "https://discord.com/api/v9/entitlements/gift-codes/"+tokens[i]+"?country_code=US&payment_source_id=750578444684754944&with_application=false&with_subscription_plan=true", nil)

			if err != nil {
				fmt.Fprintln(color.Output, red("[", currentTime(), "] Thread #"+fmt.Sprint(id+1), ":", "Invalid promo (", tokens[i], ")"))
				invalidPromos = append(invalidPromos, tokens[i])
				return
			}

			req.Header.Add("Content-Type", "application/json")
			r, err := client.Do(req)

			if err != nil {
				fmt.Fprintln(color.Output, red("[", currentTime(), "] Thread #"+fmt.Sprint(id+1), ":", "Invalid promo (", tokens[i], ")"))
				invalidPromos = append(invalidPromos, tokens[i])
				return
			}

			defer r.Body.Close()

			data, err := ioutil.ReadAll(r.Body)

			if err != nil {
				fmt.Fprintln(color.Output, red("[", currentTime(), "] Thread #"+fmt.Sprint(id+1), ":", "Invalid promo (", tokens[i], ")"))
				invalidPromos = append(invalidPromos, tokens[i])
				return
			}

			var user map[string]interface{}

			err = json.Unmarshal(data, &user)

			if err != nil {
				fmt.Fprintln(color.Output, red("[", currentTime(), "] Thread #"+fmt.Sprint(id+1), ":", "Invalid promo (", tokens[i], ")"))
				invalidPromos = append(invalidPromos, tokens[i])
				return
			}

			uses := user["uses"]

			if uses == nil {
				fmt.Fprintln(color.Output, red("[", currentTime(), "] Thread #"+fmt.Sprint(id+1), ":", "Invalid promo (", tokens[i], ")"))
				invalidPromos = append(invalidPromos, tokens[i])
				return
			}

			if uses.(float64) != 0 {
				fmt.Fprintln(color.Output, red("[", currentTime(), "] Thread #"+fmt.Sprint(id+1), ":", "Invalid promo (", tokens[i], ")"))
				invalidPromos = append(invalidPromos, tokens[i])
				return
			}

			validPromos = append(validPromos, tokens[i])
			fmt.Fprintln(color.Output, green("[", currentTime(), "] Thread #"+fmt.Sprint(id+1), ":", "Valid promo (", tokens[i], ")"))

		}(i)
	}

	thread.Wait()

	fmt.Fprintln(color.Output, cyan("[", currentTime(), "] Thread #"+fmt.Sprint(id+1)+": Finished all his tasks!"))
}

func createFile(fileName string, collection []string) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	f, err := os.Create(fileName)

	if err != nil {
		fmt.Fprintln(color.Output, red("[", currentTime(), "]  An error occurred while writing the ", fileName, " file."))
		return
	}

	defer f.Close()

	_, err2 := f.WriteString(strings.Join(collection, "\n"))

	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Fprintln(color.Output, green("[", currentTime(), "]  File ", fileName, " successfully created!"))
}

func readFile(filename string) []string {
	textCollection := []string{}
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		textCollection = append(textCollection, scanner.Text())
	}
	return textCollection
}

func currentTime() string {
	return strings.Split(time.Now().Format("15:04:05"), " ")[0]
}
