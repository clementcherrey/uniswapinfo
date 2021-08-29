package handlers

import (
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"uniswapinfo/graph"

	"github.com/gin-gonic/gin"
)

func GetPools(c *gin.Context) {
	assetID := c.Param("id")

	type graphData struct {
		Pools []struct {
			ID string `json:"id"`
		} `json:"pools"`
	}

	dataToken0, dataToken1 := new(graphData), new(graphData)
	queryToken0 := fmt.Sprintf(`{pools (where: {token0: %q}){id}}`, assetID)
	queryToken1 := fmt.Sprintf(`{pools (where: {token1: %q}){id}}`, assetID)

	var wg sync.WaitGroup
	wg.Add(2)
	go GraphWorker(&wg, c, queryToken0, dataToken0)
	go GraphWorker(&wg, c, queryToken1, dataToken1)
	wg.Wait()

	// send back both result merged
	c.JSON(http.StatusOK, gin.H{
		"pools": append(dataToken0.Pools, dataToken1.Pools...),
	})
}

func GetVolume(c *gin.Context) {
	assetID := c.Param("id")

	start, end := c.Query("start"), c.Query("end")
	if len(start) == 0 || len(end) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "missing time range",
			"message": `"start" and "end" query parameters are required`,
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "start: invalid format",
			"message": `timestamp must follow RFC3339 layout (2006-01-02T15:04:05Z07:00)`,
		})
		return
	}
	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "end: invalid format",
			"message": `timestamp must follow RFC3339 layout (2006-01-02T15:04:05Z07:00)`,
		})
		return
	}

	if !endTime.After(startTime) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid time range",
		})
		return
	}

	dataToken0 := new(struct {
		Swaps []struct {
			Amount0 string `json:"amount0"`
		} `json:"swaps"`
	})
	dataToken1 := new(struct {
		Swaps []struct {
			Amount1 string `json:"amount1"`
		} `json:"swaps"`
	})

	queryToken0 := fmt.Sprintf(`{swaps(where: {timestamp_gt: %d, timestamp_lt: %d, token0: %q }) { amount0 }}`,
		startTime.Unix(), endTime.Unix(), assetID)

	queryToken1 := fmt.Sprintf(`{swaps(where: {timestamp_gt: %d, timestamp_lt: %d, token1: %q }) { amount1 }}`,
		startTime.Unix(), endTime.Unix(), assetID)

	var wg sync.WaitGroup
	wg.Add(2)
	go GraphWorker(&wg, c, queryToken0, dataToken0)
	go GraphWorker(&wg, c, queryToken1, dataToken1)
	wg.Wait()

	sumAbs := new(big.Float)
	for _, s := range dataToken0.Swaps {
		v, _, err := new(big.Float).Parse(s.Amount0, 10)
		if err != nil {
			panic(err)
		}
		sumAbs.Add(sumAbs, v.Abs(v))
	}
	for _, s := range dataToken1.Swaps {
		v, _, err := new(big.Float).Parse(s.Amount1, 10)
		if err != nil {
			panic(err)
		}
		sumAbs.Add(sumAbs, v.Abs(v))
	}

	c.JSON(http.StatusOK, gin.H{
		"volume": sumAbs.Text('f', 18),
	})
}

func GraphWorker(wg *sync.WaitGroup, ctx *gin.Context, query string, dataStructure interface{}) {
	defer wg.Done()
	if err := graph.RunQuery(ctx, query, dataStructure); err != nil {
		log.Println(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}
