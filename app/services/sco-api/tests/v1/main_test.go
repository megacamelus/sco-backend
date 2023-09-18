package tests

import (
	"fmt"
	"testing"

	"github.com/sco1237896/sco-backend/business/data/dbtest"
	"github.com/sco1237896/sco-backend/foundation/docker"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}
