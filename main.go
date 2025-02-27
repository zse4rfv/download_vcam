package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func main() {

	url := flag.String("url", "", "url")
	user := flag.String("user", "", "user")
	pass := flag.String("pass", "", "pass")
	catalog := flag.String("catalog", "", "catalog")
	typecam := flag.String("typecam", "", "typecam")
	//	mg_email := flag.String("mgemail", "", "mega email")
	//	mg_pass := flag.String("mgpass", "", "mega pass")
	//	mg_catalog := flag.String("mgcatalog", "", "mega catalog")

	flag.Parse()

	if *url == "" || *user == "" || *pass == "" {
		log.Fatalf("Error input arguments: Needs -url=http://temp.com:999 -user=admin -pass=12345 -catalog=C:\\Temp\\ -typecam=1") //-mgemail=admin@live.com -mgpass=12345 -mgcatalog=Vs")
	}

	/* MEGA - Cloud Upload test
	email := ""
	pass_mg := ""
	cat := "Vcam"
	file_name := "20221004193413_100.h264"

	var prg *chan int
	var mc_cat *mega.Node
	var fl_ok int

	mc := mega.New()
	mc.Login(email, pass_mg)

	mc_note, err := mc.FS.GetChildren(mc.FS.GetRoot())
	if err != nil {
		log.Fatalf("Error opening MEGA: %v", err)
	}

	for i := range mc_note {
		if mc_note[i].GetName() == cat {
			mc_cat = mc_note[i]
			break
		}
	}

	if mc_cat == nil {
		mc_cat, err = mc.CreateDir(cat, mc.FS.GetRoot())
		if err != nil {
			log.Fatalf("Error opening MEGA: %v", err)
		}
	}

	mc_files, err := mc.FS.GetChildren(mc_cat)
	if err != nil {
		log.Fatalf("Error opening MEGA: %v", err)
	}

	for i := range mc_files {
		if mc_files[i].GetName() == file_name {
			fl_ok = 1
			break
		}
	}
	if fl_ok != 1 {
		m, err := mc.UploadFile(catalog+file_name, mc_cat, file_name, prg)
		if err != nil {
			log.Fatalf("Error upload to MEGA: %v", err)
		}
		println(m)
	}
	*/
	client := &http.Client{}
	tm := time.Now()

	f, err := os.OpenFile(*catalog+"logDownload"+tm.Format("010220061504")+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Println("Start")
	println(tm.Format("01/02/2006 15:04"), "Start")
	GetVideo(client, *url, *user, *pass, *catalog, *typecam)
	tm = time.Now()
	log.Println("Finish")
	println(tm.Format("01/02/2006 15:04"), "Finish")

}

func get_content_data(doc *html.Node, strtype string) []string {

	content := []string{}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" && (strings.HasPrefix(a.Val, strtype) || strings.HasSuffix(a.Val, strtype)) {
					content = append(content, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return content
}

func GetVideo(client *http.Client, url string, username string, passwd string, catalog string, typecam string) string {

	var e, w, i int
	files := make([]string, 0)
	filelenghts := make([]string, 0)

	switch typecam {
	case "1":
		{
			req, err := http.NewRequest(http.MethodGet, url+"/sd/", nil)
			req.SetBasicAuth(username, passwd)
			q := req.URL.Query()
			req.URL.RawQuery = q.Encode() // assign encoded query string to http request
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			bodyText, err := ioutil.ReadAll(resp.Body)
			s := string(bodyText)

			if strings.Contains(s, "Auth Failed") || strings.Contains(s, "Error: username or password error") {
				log.Fatal("Error - Invalid user credentials")
			}

			doc, err := html.Parse(strings.NewReader(s))
			if err != nil {
				log.Fatal(err)
			}

			content := get_content_data(doc, "/sd/20")

			for cnt := range content {
				file_str := url + content[cnt] + "record000/"
				//fmt.Println(file_str)
				req, err := http.NewRequest(http.MethodGet, file_str, nil)
				req.SetBasicAuth(username, passwd)
				q := req.URL.Query()
				req.URL.RawQuery = q.Encode() // assign encoded query string to http request
				resp, err := client.Do(req)
				if err != nil {
					log.Fatal(err)
				}
				defer resp.Body.Close()
				bodyText, err := ioutil.ReadAll(resp.Body)
				s = string(bodyText)
				if strings.Contains(s, "Auth Failed") || strings.Contains(s, "Error: username or password error") {
					log.Fatal("Error - Invalid user credentials")
				}

				doc, err := html.Parse(strings.NewReader(s))
				if err != nil {
					log.Fatal(err)
				}

				contentf := get_content_data(doc, ".265")
				for cntf := range contentf {
					//fmt.Println(contentf[cntf])
					files = append(files, contentf[cntf])
				}
				contentf = get_content_data(doc, ".264")
				for cntf := range contentf {
					//fmt.Println(contentf[cntf])
					files = append(files, contentf[cntf])
				}
			}
		}
	default:
		{
			req, err := http.NewRequest(http.MethodGet, url+"/get_record_file.cgi?PageSize=10000", nil)
			q := req.URL.Query()
			q.Add("loginuse", username)
			q.Add("loginpas", passwd)
			req.URL.RawQuery = q.Encode() // assign encoded query string to http request
			resp, err := client.Do(req)
			//resp, err := http.Get(url + "/get_record_file.cgi" + "?loginuse=" + username + "&loginpas=" + passwd + "&PageSize=10000")
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			bodyText, err := ioutil.ReadAll(resp.Body)
			s := string(bodyText)

			if strings.Contains(s, "Auth Failed") {
				log.Fatal("Error - Invalid user credentials")
			}

			lines := strings.Split(s, "\n")

			for line := range lines {
				s := lines[line]

				index := strings.Index(s, "h264")
				if index > 0 {
					files = append(files, s[(index-19):index+4])
				}
				index = strings.Index(s, "record_size0[")
				if index >= 0 {
					first := index + 16
					last := index + 24
					for i := index; i < len(s); i++ {
						smb := string(s[i])
						if smb == "=" {
							first = i + 1
						} else if smb == ";" {
							last = i
						}
					}
					filelenghts = append(filelenghts, s[first:last])
				}
			}

		}
	}

	sort.Sort(sort.StringSlice(files))
	var fileurl, str_file string

	for file := range files {
		switch typecam {
		case "1":
			{
				str_http := "http://"
				fileurl = str_http + username + ":" + passwd + "@" + strings.TrimPrefix(url, str_http) + files[file]
				str_file = strings.TrimRight(files[file], "/")
				str_file = strings.Split(str_file, "/")[len(strings.Split(str_file, "/"))-1]
			}
		default:
			{
				fileurl = url + "/record/" + files[file] + "?loginuse=" + username + "&loginpas=" + passwd
				str_file = files[file]
			}
		}
		tm := time.Now()
		log.Println(len(files), file+1, str_file) //, "size:", filelenghts[file])
		print(tm.Format("01/02/2006 15:04 "), len(files), "/", file+1, " ", str_file)
		r := downloadfile(client, fileurl, str_file, catalog)
		switch r {
		case 0:
			e++
			println(" error")
		case 1:
			i++
			println(" ok")
		case 2:
			w++
			println(" skip")
		}
	}
	log.Println("Download", i, "Skip", w, "Error", e)
	println("Download", i, "Skip", w, "Error", e)
	return "Ok"
}

func downloadfile(client *http.Client, fileurl, file, catalog string) int {

	if _, err := os.Stat(catalog + file); errors.Is(err, os.ErrNotExist) {
		//OK
	} else {
		log.Println("File already exist, skip")
		return 2
	}

	req, err := http.NewRequest(http.MethodGet, fileurl, nil)
	resp, err := client.Do(req)
	//resp, err := http.Get(fileurl)
	if err != nil {
		log.Println("Error get file, skip")
		//log.Fatal(err)
		return 0
	}
	defer resp.Body.Close()

	dwn, err := ioutil.ReadAll(resp.Body)
	if err != nil || len(dwn) <= 0 {
		log.Println("Error read file, skip")
		//log.Fatal(err)
		return 0
	}

	s := string(dwn)
	if strings.Contains(s, "Error: Unauthorized") {
		log.Println("Error: Invalid user credentials, skip")
		return 0
	}

	er := ioutil.WriteFile(catalog+file, dwn, 0444)
	if er != nil {
		log.Println("Error write file, skip")
		//log.Fatal(err)
		return 0
	}

	log.Println("File was download successful, size: ", len(dwn))
	return 1
}
