package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Find - глобальный флаг успешного нахождения той самой
var Find = false

const (
	// DOSecretToket - секретный ключ digital ocean
	DOSecretToket = "XXXXXXXXXXXXXXXX"
)

// NewDropletParams - структура нового дроплета
type NewDropletParams struct {
	Name   string   `json:"name"`
	Region string   `json:"region"`
	Size   string   `json:"size"`
	Image  string   `json:"image"`
	Tags   []string `json:"tags"`
}

// Droplet - данные дроплета
type Droplet struct {
	Data struct {
		ID       int      `json:"id"`
		Status   string   `json:"status"`
		Networks Networks `json:"networks"`
	} `json:"droplet"`
}

// Networks - network дроплета
type Networks struct {
	V4 []NetworkV4 `json:"v4"`
}

// NetworkV4 - ipv4 дроплета
type NetworkV4 struct {
	IP string `json:"ip_address"`
}

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// go startSearchAvailableIP()
	// go startSearchAvailableIP()
	// go startSearchAvailableIP()
	// go startSearchAvailableIP()
	// go startSearchAvailableIP()
	// go startSearchAvailableIP()
	// go startSearchAvailableIP()
	// go startSearchAvailableIP()
	// go startSearchAvailableIP()

	deleteAllTmpDroplets()

	<-signals
	close(signals)
	os.Exit(0)
}

func startSearchAvailableIP() {
	var pingCheck bool
	log.Println("Start create droplets, while not found available ip")

	for !pingCheck {
		log.Println("=====================================================")
		id, err := addDroplet()
		if err != nil {
			log.Fatalf("Add droplet error: %v", err)
		}

		ip, err := getDropletIP(id)
		if err != nil {
			log.Fatalf("Get droplet ip error: %v", err)
		}

		pingCheck, err = checkPing(ip)
		if err != nil {
			log.Fatalf("Ping check failed: %v", err)
		}

		if !pingCheck || Find {
			if err := deleteDroplet(id); err != nil {
				log.Fatalf("Delete droplet failed: %v", err)
			}

			return
		} else {
			log.Println("Find droplet, which ip is not blocked. exit.")
			Find = true
		}
	}
}

func addDroplet() (int, error) {
	var droplet Droplet
	defaultHTTPClient := http.Client{}
	params := &NewDropletParams{"example", "lon1", "s-1vcpu-1gb", "ubuntu-16-04-x64", []string{"tmp"}}

	b, err := json.Marshal(params)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.digitalocean.com/v2/droplets", bytes.NewReader(b))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", DOSecretToket))
	req.Header.Add("Content-type", "application/json")

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return 0, err
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if err = json.Unmarshal(b, &droplet); err != nil {
		return 0, err
	}

	log.Printf("Added droplet successfuly, id = %d", droplet.Data.ID)
	return droplet.Data.ID, nil
}

func getDropletIP(dropletID int) (string, error) {
	defaultHTTPClient := http.Client{}
	var droplet Droplet

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.digitalocean.com/v2/droplets/%d", dropletID), strings.NewReader(""))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", DOSecretToket))

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err = json.Unmarshal(b, &droplet); err != nil {
		return "", err
	}

	if len(droplet.Data.Networks.V4) == 0 || droplet.Data.Status != "active" {
		log.Printf("Get droplet ip complete, but droplet isn`t ready (%s). Try to send another request", droplet.Data.Status)
		time.Sleep(1000 * time.Millisecond)
		return getDropletIP(dropletID)
	}

	if droplet.Data.Status == "active" {
		// For sure droplet is active
		time.Sleep(3000 * time.Millisecond)
	}

	log.Printf("Get droplet ip complete, ip = %s", droplet.Data.Networks.V4[0].IP)
	return droplet.Data.Networks.V4[0].IP, nil
}

func deleteDroplet(id int) error {
	defaultHTTPClient := http.Client{}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://api.digitalocean.com/v2/droplets?id=%d", id), strings.NewReader(""))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", DOSecretToket))

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		log.Printf("Delete droplet failed, status = %s", resp.Status)
		return errors.New("Delete droplet failed, status != 204")
	}

	log.Printf("Delete droplet %d complete", id)
	return nil
}

func deleteAllTmpDroplets() error {
	defaultHTTPClient := http.Client{}

	req, err := http.NewRequest(http.MethodDelete, "https://api.digitalocean.com/v2/droplets?tag_name=tmp", strings.NewReader(""))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", DOSecretToket))

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		log.Printf("Delete droplet failed, status = %s", resp.Status)
		return errors.New("Delete droplet failed, status != 204")
	}

	log.Printf("Delete all tmp droplets complete, exit.")
	os.Exit(0)
	return nil
}

func checkPing(ip string) (bool, error) {
	out, _ := exec.Command("ping", ip, "-c 2").Output()

	if strings.Contains(string(out), ", 0% packet loss") {
		log.Printf("Check ping complete, no packet loss")
		return true, nil
	} else {
		log.Printf("Check ping complete, some packets loss")
		return false, nil
	}
}
