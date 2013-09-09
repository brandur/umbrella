package main

import "fmt"
import "os"

func RequireEnv(key string) string {
  value := os.Getenv(key)
  if value == "" {
    panic(fmt.Sprintf("missing=%s\n", key))
  }
  return value
}
