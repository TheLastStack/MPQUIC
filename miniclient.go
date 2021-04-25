package main

import (
  "crypto/tls"
  "fmt"
  "io"
  "os/exec"
  "flag"
  "strings"
  "github.com/lucas-clemente/quic-go"
)

func errored(err error) {
  if err != nil {
    panic(err)
  }
}

func check_server_comm(a string, b string) {
  if strings.Compare(a, b) != 0 {
    panic("Client disconnected")
  }
}

func main() {
  var host_address string
  var host_port string
  var file_name string
  flag.StringVar(&host_address, "ip", "localhost", "IP address where server is to listen to")
  flag.StringVar(&host_port, "port", "8000", "Port number at which server will listen")
  flag.StringVar(&file_name, "file", "Udemy", "File to be streamed")
  flag.Parse()
  var host_tuple string
  host_tuple = host_address + ":" + host_port
  quicConfig := &quic.Config{
    CreatePaths: true,
  }
  server_conn, err := quic.DialAddr(host_tuple, &tls.Config{InsecureSkipVerify: true}, quicConfig)
  errored(err)
  stream, err := server_conn.OpenStreamSync()
  send(stream, "READY", ".")
  response := read_until(stream, ".")
  check_server_comm(response, "SEND")
  send(stream, file_name, ".")
  response = read_until(stream, ".")
  check_server_comm(response, "PRESENT")
  response = read_until(stream, ".")
  duration := strings.Replace(response, ",", ".", 1)
  fmt.Println("Duration is: " + duration)
  send(stream, "ACK", ".")
  fmt.Println("Started streaming")
  ffmpeg := exec.Command("ffmpeg", "-i", "pipe:", "-c", "copy", "video.ts")
  inpipe, err := ffmpeg.StdinPipe()
  errored(err)
  err = ffmpeg.Start()
  _, err = io.Copy(inpipe, stream)
  fmt.Println("Terminated Streaming")
  ffmpeg.Wait()
}

func substr(input string, start int, length int) string {
    asRunes := []rune(input)
    
    if start >= len(asRunes) {
        return ""
    }
    
    if start+length > len(asRunes) {
        length = len(asRunes) - start
    }
    
    return string(asRunes[start : start+length])
}

func read_until(stream quic.Stream, delimiter string) string {
  len_bytes := len([]byte("1"))
  complete_data := ""
  read_data := make([]byte, len_bytes)
  for string(read_data) != delimiter {
    stream.Read(read_data)
    complete_data = complete_data + string(read_data)
  }
  return substr(complete_data, 0, len([]rune(complete_data)) - 1)
}

func send(stream quic.Stream, message string, delimiter string) {
  stream.Write([]byte(message + delimiter))
}
