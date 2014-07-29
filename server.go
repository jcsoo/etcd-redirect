package main
import (
  "log"
  "strings"
  "regexp"
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
  key := "/config/redirects/"+r.Host
  response, err := client.Get(key, false, false)
  if err != nil {
    log.Printf("Error looking up host: %s\n",err)
    http.NotFound(w, r)
    return
  }
  lines := strings.Split(response.Node.Value,"\n")
  for _, line := range lines {
    fields := strings.Split(line," ")
    if len(fields) < 1 {
      continue
    }
    pattern := fields[0]
    template := fields[1]
    re, err := regexp.Compile(pattern)
    if err != nil {
      log.Printf("Error compiling regular expression %s: %s", pattern, err)
      continue
    }
    if re.MatchString(r.RequestURI) == true {
      target := []byte{}
      for _, s := range re.FindAllStringSubmatchIndex(r.RequestURI,-1) {
        target = re.ExpandString(target, template, r.RequestURI, s)
	    }

      log.Printf("%s matched URI %s", pattern, r.RequestURI)
      log.Printf("http://%s%s -> %s\n", r.Host, r.RequestURI, target)
      http.Redirect(w, r, string(target) ,302)
      return
    }
  }
  http.NotFound(w, r)
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":5000", nil)
}
