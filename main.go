package main

/*
#cgo pkg-config: python-3.9
#cgo linux LDFLAGS:  -lpython3.7m
#cgo linux CFLAGS: -I/usr/include/python3.7m
#include "lib_c.h"
*/
import "C"
import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	"os/signal"
    "syscall"

	"github.com/polarspetroll/gopio"
)

var (
	total_consumption float64
	INTERVALSLEEP     time.Duration = 657 * time.Millisecond
	mutex             sync.Mutex
	relay_pins        []int              = []int{}
	sessions          []SID              = []SID{}
	trials            []User             = []User{}
	tmps              *template.Template = template.Must(template.ParseGlob("templates/*.gohtml"))
)

type HTML struct {
	Username string
	Message  string
}

type Trial struct {
	TimeLeft  time.Duration
	UnitsLeft float64
	Price     int
}

type User struct {
	Username  string
	Password  string
	Condition Trial
	RelayPin  gopio.WiringPiPin
	InaAddr   int
}

type SID struct {
	Sid      string
	Username string
}

func main() {
	gopio.GopioSetUp()
	ParseConfig()
	var pin gopio.WiringPiPin
	for _, p := range relay_pins {
		pin = gopio.PinMode(p, gopio.OUT)
		pin.DigitalWrite(gopio.LOW)
	}

	var usr User
	files, _ := ioutil.ReadDir("./Database")
	for _, v := range files {
		usr, _ = GetUser(strings.ReplaceAll(v.Name(), ".json", ""))
		if (usr.Condition == Trial{}) {
			continue
		}
		trials = append(trials, usr)
	}

	http.HandleFunc("/", Index)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/signup", SignUp)
	http.Handle("/statics/", http.StripPrefix("/statics/", http.FileServer(http.Dir("statics"))))

	go cookieInterval()
	go func() {
		for {
			for i, v := range trials {
				go TrialInterval(v)
				RemoveTrial(i)
			}
		}
	}()
	go Monitor(len(relay_pins))
	go Listen()
	http.ListenAndServe(":8080", nil)
}

func Index(w http.ResponseWriter, r *http.Request) {

	st, username := CheckCookie(r)
	if !st {
		http.Redirect(w, r, "/login", 302)
		return
	}
	user, err := GetUser(username)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if r.Method == http.MethodGet {
		tmps.ExecuteTemplate(w, "index.gohtml", HTML{Username: username, Message: ""})
	} else if r.Method == http.MethodPost {
		kw := r.PostFormValue("kw")
		timeUnit := r.PostFormValue("unit")
		p := r.PostFormValue("period")
		kwatt, err := strconv.ParseFloat(kw, 64)

		if err != nil {
			http.Error(w, "Invalid Duration Or Consumption", 500)
			return
		}

		var total time.Duration
		switch timeUnit {
		case "hour":
			total, _ = time.ParseDuration(p + "h")
			break
		case "day":
			total = ParseTimeHour(24)
			break
		case "month":
			total = ParseTimeHour(24 * 30)
			break
		case "week":
			total = ParseTimeHour(24 * 7)
			break
		case "minute":
			total, _ = time.ParseDuration(fmt.Sprintf("%vm", p))
			break
		case "second":
			total, _ = time.ParseDuration(fmt.Sprintf("%vs", p))
			break
		default:
			http.Error(w, "Invalid Duration", 500)
			return
		}
		relay_len := len(relay_pins)
		if relay_len == 0 {
			http.Error(w, "Out Of Service", 500)
			return
		}
		price := CalculatePrice(total, int(kwatt))
		indx := len(relay_pins) - 1
		user.RelayPin = gopio.PinMode(relay_pins[indx], gopio.OUT)
		relay_pins = append(relay_pins[:indx], relay_pins[indx+1:]...)
		new_trial := Trial{TimeLeft: total, Price: price, UnitsLeft: kwatt}
		user.NewTrial(new_trial)
		trials = append(trials, user)
		tmps.ExecuteTemplate(w, "index.gohtml", HTML{Username: username, Message: "Done!"})
	}

}

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmps.ExecuteTemplate(w, "signup.gohtml", nil)
		return
	} else if r.Method == http.MethodPost {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		_, err := GetUser(username)
		if err == nil {
			tmps.ExecuteTemplate(w, "signup.gohtml", "User Exists")
			return
		}
		if NewUser(User{Username: username, Password: password}) != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		cookie := GetNewCookie(username)
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", 302)
	} else {
		http.Error(w, "Method Not Allowed", 405)
		return
	}

}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmps.ExecuteTemplate(w, "login.gohtml", nil)
	} else if r.Method == http.MethodPost {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		user, err := GetUser(username)
		if err != nil || user.Password != password {
			tmps.ExecuteTemplate(w, "login.gohtml", "Invalid username or password")
			return
		}
		cookie := GetNewCookie(user.Username)
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", 302)
	} else {
		http.Error(w, "Method Not Allowed", 405)
	}
}

func CheckCookie(r *http.Request) (bool, string) {
	c, err := r.Cookie("P2PSSID")
	if err != nil {
		return false, ""
	}
	for _, v := range sessions {
		if v.Sid == c.Value {
			return true, v.Username
		}
	}
	return false, ""
}

func CalculatePrice(p time.Duration, kw int) (price int) {
	p = p / time.Second
	price = kw * 100 // example calculation function for consumption
	return price
}

func NewUser(user User) error {
	f, err := os.Create("Database/" + user.Username + ".json")
	defer f.Close()
	if err != nil {
		return err
	}
	j, err := json.MarshalIndent(user, "", " ")
	if err != nil {
		return err
	}
	f.Write(j)
	return nil
}

func GetUser(username string) (User, error) {
	f, err := os.Open("Database/" + username + ".json")
	defer f.Close()
	if err != nil {
		return User{}, err
	}
	r, err := ioutil.ReadAll(f)
	if err != nil {
		return User{}, err
	}
	var user User
	err = json.Unmarshal(r, &user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (u *User) NewTrial(t Trial) error {
	u.Condition = t
	f, err := os.Open("Database/" + u.Username + ".json")
	defer f.Close()
	if err != nil {
		return err
	}
	j, err := json.MarshalIndent(u, "", " ")
	if err != nil {
		return err
	}
	f.Write(j)
	return nil
}

func (u *User) EditTrial(t Trial) error {
	u.Condition = t
	f, err := os.Open("Database/" + u.Username + ".json")
	defer f.Close()
	if err != nil {
		return err
	}
	j, err := json.MarshalIndent(u, "", " ")
	if err != nil {
		return err
	}
	f.Write(j)
	return nil
}

func (u *User) DisconnectPower() {
	u.RelayPin.DigitalWrite(gopio.LOW)
}

func (u *User) ConnectPower() {
	u.RelayPin.DigitalWrite(gopio.HIGH)
}

func TrialInterval(u User) {
	u.ConnectPower()
	for u.Condition.TimeLeft >= 0 && u.Condition.UnitsLeft >= 0 {
		u.AddConsumption()
		u.Condition.TimeLeft -= INTERVALSLEEP
		u.EditTrial(Trial{TimeLeft: u.Condition.TimeLeft, UnitsLeft: u.Condition.UnitsLeft, Price: u.Condition.Price})
		time.Sleep(INTERVALSLEEP)

	}

	u.DisconnectPower()
	relay_pins = append(relay_pins, u.RelayPin.Num) // free the pin
	u.RelayPin = gopio.WiringPiPin{}
}

func RemoveTrial(index int) {
	trials = append(trials[:index], trials[index+1:]...)
}

func GetNewCookie(username string) http.Cookie {
	var a []byte = make([]byte, 20)
	rand.Read(a)
	sid := fmt.Sprintf("%x", a)
	sessions = append(sessions, SID{Sid: sid, Username: username})
	return http.Cookie{Name: "P2PSSID", Value: sid, HttpOnly: true}
}

func ParseTimeHour(i int64) time.Duration {
	t, _ := time.ParseDuration(fmt.Sprintf("%dns", i*int64(time.Hour)))
	return t
}

////////////////////////

func cookieInterval() {
	for {
		sessions = []SID{}
		time.Sleep(24 * time.Hour)
	}
}

func Monitor(m int) {
	C.LCD_Write(C.CString("Service Started!"), C.int(1), C.CString("left"))

	time.Sleep(3 * time.Second)
	var val string
	var val2 string
	var length int
	var temp string
	var out []byte
	for {
		length = len(relay_pins)
		if length == 0 {
			C.LCD_Write(C.CString("Out Of Service"), C.int(1), C.CString("left"))
			time.Sleep(2 * time.Second)
		} else {
			val = fmt.Sprintf("Total Relays:%d", m)
			val2 = fmt.Sprintf("Free Relays:%d", length)
			C.Double_Write(C.CString(val), C.CString(val2))
			time.Sleep(3 * time.Second)
		}
		C.LCD_Write(C.CString(fmt.Sprintf("Consumption:%v", total_consumption)), C.int(1), C.CString("left"))
		time.Sleep(3 * time.Second)
		C.LCD_Clear()
		out, _ = exec.Command("/opt/vc/bin/vcgencmd", "measure_temp").Output()
		temp = string(out)
		temp = strings.ReplaceAll(temp, "temp=", "")
		temp = strings.ReplaceAll(temp, "'C\n", "c")
		C.LCD_Write(C.CString(fmt.Sprintf("CPU Temp:%s", temp)), C.int(1), C.CString("left"))
		time.Sleep(2 * time.Second)
	}
}

type Config struct {
	Pins []int `json:"relays"`
}

func ParseConfig() {
	f, err := os.Open("config.json")
	defer f.Close()
	if err != nil {
		log.Fatal("Config file not found")
	}

	b, _ := ioutil.ReadAll(f)
	var out Config
	json.Unmarshal(b, &out)
	for _, v := range out.Pins {
		relay_pins = append(relay_pins, v)
	}
}


/// SIGNAL ///
func Listen() {

    sigs := make(chan os.Signal, 1)
    done := make(chan bool, 1)

    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        sig := <-sigs
        fmt.Printf("New OS Signal : %v\n", sig)
        done <- true
    }()

    <-done
    func(){
    	var pin gopio.WiringPiPin
	    for _, p := range relay_pins {
			pin = gopio.PinMode(p, gopio.OUT)
			pin.DigitalWrite(gopio.LOW)
		}
		os.Exit(0)
	}()
}