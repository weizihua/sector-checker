package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	// go test -timeout 30s sector-checker -run ^TestTime

	// Running tool: /usr/local/go/bin/go test -coverprofile=/var/folders/v3/htc95bpx66g15z5tqrm3d2800000gn/T/go-code-cover -timeout 30s -run ^TestTime$

	windowpostStart := time.Now()
	// time.Sleep(time.Duration(10) * time.Second)

	windowpost1 := windowpostStart.Add(time.Duration(3*60) * time.Second)

	// windowpost1 := time.Now()
	PostGenerateCandidates := windowpost1.Sub(windowpostStart)

	verifyWindowpost1 := windowpost1.Add(time.Duration(5*60) * time.Second)

	bo := CheckResults{}

	bo.PostGenerateCandidates = windowpost1.Sub(windowpostStart)
	bo.VerifyWinningPostCold = verifyWindowpost1.Sub(windowpost1)

	// bo.PostGenerateCandidatesM = bo.PostGenerateCandidates.Truncate(time.Millisecond * 100)
	// bo.VerifyWinningPostColdM = bo.VerifyWinningPostCold.Truncate(time.Millisecond * 100)

	fmt.Printf("bo == %+v:\n", bo)
	fmt.Printf("PostGenerateCandidates == %+v:\n", bo.PostGenerateCandidates)
	fmt.Printf("VerifyWinningPostCold == %+v:\n", bo.VerifyWinningPostCold)
	fmt.Printf("PostGenerateCandidatesM == %+v:\n", bo.PostGenerateCandidatesM)
	fmt.Printf("VerifyWinningPostColdM == %+v:\n", bo.VerifyWinningPostColdM)

	data, err := json.MarshalIndent(bo, "", "  ")
	if err != nil {

		fmt.Println("err == :", err)
	}

	fmt.Println(string(data))

	// t.Logf("windowpostStart: %v, windowpost1: %v, PostGenerateCandidates: %v", windowpostStart, windowpost1, PostGenerateCandidates)
	t.Log(windowpostStart)
	t.Log(windowpost1)
	t.Log(PostGenerateCandidates)
}
