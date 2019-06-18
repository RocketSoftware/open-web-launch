package messaging

import (
	"os"
	"testing"
)

func TestStatusMessage(t *testing.T) {
	message := `{"status": "get"}`
	file, err := os.Create("status_message.dat")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := SendMessage(file, message); err != nil {
		panic(err)
	}
}

func TestSendMessage(t *testing.T) {
	message := `{"jnlp": "https://docs.oracle.com/javase/tutorialJWS/samples/deployment/dynamictree_webstartJWSProject/dynamictree_webstart.jnlp"}`
	file, err := os.Create("jnlp_message.dat")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := SendMessage(file, message); err != nil {
		panic(err)
	}
}
