package main
import (
  "crypto/rand"
  "crypto/rsa"
  "crypto/tls"
  "crypto/x509"
  "encoding/pem"
  "io/ioutil"
  "fmt"
  "flag"
  "math/big"
  "os"
  "io"
  "os/exec"
  "bytes"
  "strings"
  "path/filepath"
  quic "github.com/lucas-clemente/quic-go"
)
func errored(err error) {
  if err != nil {
    panic(err)
  }
}

func check_client_comm(a string, b string) {
  if strings.Compare(a, b) != 0 {
    panic("Client disconnected")
  }
}

func main() {
  var host_address string
  var host_port string
  flag.StringVar(&host_address, "ip", "localhost", "IP address where server is to listen to")
  flag.StringVar(&host_port, "port", "8000", "Port number at which server will listen")
  flag.Parse()
  var host_tuple string
  host_tuple = host_address + ":" + host_port
  fmt.Println("Server running at ", host_tuple)
  quicConfig := &quic.Config{
    CreatePaths: true,
  }
  sock, err := quic.ListenAddr(host_tuple, generateTLSConfig(), quicConfig)
  errored(err)
  dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
  errored(err)
  for {
    client_con, err := sock.Accept()
    errored(err)
    fmt.Println("Accpted client connection")
    go handle_client(client_con, dir)
  }
}

func handle_client(client_conn quic.Session, dir string) {
  stream, err := client_conn.AcceptStream()
  defer stream.Close()
  errored(err)
  client_comm := read_until(stream, ".")
  check_client_comm(client_comm, "READY")
  send(stream, "SEND", ".")
  filename := "/home/mininet/Downloads/SampleVideos/" + read_until(stream, ".") + ".mp4"
  file, err := ioutil.ReadFile(filename)
  errored(err)
  if err == nil {
    send(stream, "PRESENT", ".")
  } else {
    send(stream, "NOTPRESENT", ".")
  } 
  duration_cmd := exec.Command("ffprobe", "-sexagesimal", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)
  duration_out, err := duration_cmd.StdoutPipe()
  errored(err)
  err = duration_cmd.Start()
  outC := make(chan string)
  go func() {
    var buf bytes.Buffer
    io.Copy(&buf, duration_out)
    outC <- buf.String()
    }()
  duration_cmd.Wait()
  duration := <-outC
  fmt.Println(duration)
  send(stream, strings.Replace(duration, ".", ",", 1), ".")
  ack := read_until(stream, ".")
  check_client_comm(ack, "ACK")
  fmt.Println("Opening file:", filename)
  reader := bytes.NewReader(file)
  _, err = io.Copy(stream, reader)
  errored(err)
  fmt.Println("Completed streaming")
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
  len_byte := len([]byte("1"))
  read_data := make([]byte, len_byte)
  complete_data := ""
  for string(read_data) != delimiter {
    stream.Read(read_data)
    complete_data = complete_data + string(read_data)
  }
  return substr(complete_data, 0, len([]rune(complete_data)) - 1)
}

func send(stream quic.Stream, message string, delimiter string) {
  stream.Write([]byte(message + delimiter))
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
