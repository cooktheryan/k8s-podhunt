package main

import (
	"fmt"
	"math/rand"
	"io"
	"log"
	"net/http"
	"os"
)

func handleKill(w http.ResponseWriter, r *http.Request) {
	_, err := io.Copy(os.Stdout, r.Body)
	if err != nil {
		log.Println("Failed to parse the request")
		log.Println(err)
		return
	}

	c, err := inClusterLogin()
	if err != nil {
		log.Println("Failed to login in cluster")
		log.Println(err)
		return
	}

	message := ""

	switch rand.Intn(killOptions) {
	case 0:
		message, err = killRandomPod(c)
	case 1:
		message, err = killRandomDeployment(c)
	case 2:
		//killRandomStatefulSet(c)
	case 3:
		//updateRandomDeployement(c)
	case 4:
		//updateRandomStatefulSet(c)
	case 5:
		//killRandomWorker(c)
	}
	if err != nil {
		fmt.Fprintf(w, "{'message': '%s'}", message)
	}
}

func main() {
	log.Println("server started")
	http.HandleFunc("/kill", handleKill)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
