# Uniswap-Info

REST app in Golang that uses The Graph’s GraphQL API to provide Uniswap v3.

## 1. How to launch the app ?

Build and run the binary

```
go build uniswapinfo

./uniswapinfo
```

## 2. Usage

Get the existing pools for an asset

```
curl http://localhost:8081/asset/0x4fabb145d64652a948d72533023f6e7a623c7c53/pools
```

Get the total volume swap for an asset in a time range
```
curl http://localhost:8081/asset/0x6b175474e89094c44da98b954eedeac495271d0f/volume\?start\=2021-07-02T15:04:05Z\&end\=2021-08-20T15:04:05-07:00
```

Get the swaps that occured during a specific block
```
curl http://localhost:8081/block/12774522/swaps 
```
