package main
import (
  "log"
  "strings"
  "fmt"
  "net"
  "net/http"
  "github.com/coreos/go-etcd/etcd"
)

func checkErr(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func lookupPeersByDomain(domain string) (peers []string) {
  _, addrs, err := net.LookupSRV("etcd","tcp",domain)
  checkErr(err)
  for _, addr := range addrs {
    hostname := strings.TrimRight(addr.Target,".")
    peers = append(peers, fmt.Sprintf("http://%s:%d",hostname, addr.Port))
  }
  return
}

func handler(w http.ResponseWriter, r *http.Request) {
  peers := lookupPeersByDomain("lightreading.systems")
  client := etcd.NewClient(peers)
  key := "/config/redirects/"+r.Host+"/"+"_"
  response, err := client.Get(key, false, false)
  if err == nil {
    target := strings.Replace(response.Node.Value,"$1",r.RequestURI,1)
    log.Printf("http://%s%s -> %s\n", r.Host, r.RequestURI, target)
    http.Redirect(w, r, "http://www.google.com/",302)
  } else {
    log.Println(err)
    http.NotFound(w, r)
  }
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":5000", nil)
}
