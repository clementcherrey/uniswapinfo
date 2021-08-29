package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"uniswapinfo/graph"

	"github.com/gin-gonic/gin"
)

func GetSwaps(c *gin.Context) {
	block := c.Param("number")
	blockNumber, err := strconv.Atoi(block)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid block number",
		})
		return
	}

	// get the transaction for in the block
	txQuery := fmt.Sprintf(`{transactions (where: {blockNumber: %d}){id}}`, blockNumber)
	txData := new(struct {
		Transactions []struct {
			ID string `json:"id"`
		} `json:"transactions"`
	})
	if err := graph.RunQuery(c, txQuery, txData); err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ids := make(chan string)
	done := make(chan bool)

	var swaps []string
	go func() {
		for {
			id, more := <-ids
			if more {
				swaps = append(swaps, id)
			} else {
				done <- true
			}
		}
	}()

	var wg sync.WaitGroup
	for _, tx := range txData.Transactions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			swQuery := fmt.Sprintf(`{swaps (where: {transaction: %q}){id}}`, tx.ID)
			swData := new(struct {
				Swaps []struct {
					ID string `json:"id"`
				} `json:"swaps"`
			})

			if err := graph.RunQuery(c, swQuery, swData); err != nil {
				log.Println(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			for _, s := range swData.Swaps {
				ids <- s.ID
			}
		}()
	}
	wg.Wait()
	close(ids)

	<-done
	c.JSON(http.StatusOK, gin.H{
		"swaps": swaps,
	})
}
