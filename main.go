package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//globals------------------------------------------//
var maxNodes int = 10 /*0-9*/
var statusLog string = "Log: "
var thisNodeID int = 5
var activeNodesCount int = 1

var nodeList [10]nodeObj

var activeNodes = make([]bool, maxNodes)

var nodesMap map[int]int
var nodesTaskGroup []int

var taskFragmentLen int = 0
var HAprocessRunning bool = false

var globalResult int = 0

///THEME
var thisTheme string = "Indigo"
var themePack map[string]string

///THEME

//-------------------------------------------------//

//structures---------------------------------------//
type webPage struct {
	Title string
}

type nodeObj struct {
	NodeID           int    `json:"ID"`
	NodeCount        int    `json:"nodeCount"`
	TaskStatus       string `json:"taskStatus"` //empty-loaded-running-completed
	TaskResult       string `json:"taskResult"`
	TaskWord         string `json:"taskWord"`
	TaskFragmentBody string `json:"taskFragmentBody"`
	TaskBody         string `json:"taskBody"`
	HasAddTask       int    `json:"hasAddTask"` //ID of another node that's task now running
}

type echoMessage struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type taskMessage struct {
	TaskWord string `json:"taskWord"`
	TaskBody string `json:"taskBody"`
}

type resultMessage struct {
	TaskResult string `json:"taskResult"`
}

//-------------------------------------------------//

//global structures--------------------------------//
var thisNode nodeObj

//-------------------------------------------------//

//main func----------------------------------------//
func main() {
	///THEME
	initTheme()

	///THEME

	//current node inits
	thisNode.NodeID = thisNodeID
	thisNode.NodeCount = activeNodesCount
	thisNode.TaskStatus = "empty"
	thisNode.TaskResult = "empty"
	thisNode.TaskWord = "empty"
	thisNode.TaskFragmentBody = "empty"
	thisNode.TaskBody = "empty"
	thisNode.HasAddTask = 999

	go updateNodesStatuses()

	http.HandleFunc("/", startPage)
	http.HandleFunc("/log", logServer)
	http.HandleFunc("/clean", cleanLogs)
	http.HandleFunc("/message", message)
	http.HandleFunc("/task", taskServer)
	http.HandleFunc("/status", statusServer)
	http.HandleFunc("/updinfo", updateNodesInfo)
	http.HandleFunc("/messenger", messengerPage)
	http.HandleFunc("/test", test)
	http.HandleFunc("/reset", resetState)
	http.HandleFunc("/theme", themeServer)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

//-------------------------------------------------//

//web handlers-------------------------------------//
func message(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	node := r.FormValue("nodeId")
	nodeID, _ := strconv.Atoi(node)
	message := r.FormValue("echoContent")
	go sendUserMessage(nodeID, message)
}

func resetState(w http.ResponseWriter, r *http.Request) {
	thisNode.NodeID = thisNodeID
	thisNode.NodeCount = activeNodesCount
	thisNode.TaskStatus = "empty"
	thisNode.TaskResult = "empty"
	thisNode.TaskWord = "empty"
	thisNode.TaskFragmentBody = "empty"
	thisNode.TaskBody = "empty"
	thisNode.HasAddTask = 999
	http.Redirect(w, r, "http://goappnode"+strconv.Itoa(thisNodeID)+".herokuapp.com", http.StatusSeeOther)
}

func updateNodesInfo(w http.ResponseWriter, r *http.Request) {
	activeNodesCount = 1
	go updateNodesStatuses()
	http.Redirect(w, r, "http://goappnode"+strconv.Itoa(thisNodeID)+".herokuapp.com", http.StatusSeeOther)
}

func startPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		templatePage, _ := template.ParseFiles(themePack[thisTheme])
		templatePage.Execute(w, &webPage{"simplePage"})
	case "POST":

	}
}

func themeServer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		r.ParseForm()
		thisTheme = r.FormValue("color")
		http.Redirect(w, r, "http://goappnode"+strconv.Itoa(thisNodeID)+".herokuapp.com", http.StatusSeeOther)
	}
}

func messengerPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		templatePage, _ := template.ParseFiles("messenger.html")
		templatePage.Execute(w, &webPage{"simplePage"})
	}
}

func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Test started")
	remapNodes()
	runTask()
	http.Redirect(w, r, "http://goappnode"+strconv.Itoa(thisNodeID)+".herokuapp.com", http.StatusSeeOther)
}

func taskServer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		switch r.Header.Get("getnodeparams") {
		case "taskstatus":
			fmt.Fprint(w, thisNode.TaskStatus)
		case "taskresult":
			fmt.Fprint(w, thisNode.TaskResult)
		case "taskword":
			fmt.Fprint(w, thisNode.TaskWord)
		case "taskfragmentbody":
			fmt.Fprint(w, thisNode.TaskFragmentBody)
		case "tasktody":
			fmt.Fprint(w, thisNode.TaskBody)
		case "addtask":
			fmt.Fprint(w, thisNode.HasAddTask)
		}
	case "POST":
		switch r.Header.Get("reqtype") {
		case "taskcontent":
			thisNode.TaskStatus = "loaded"
			fromNode, _ := strconv.Atoi(r.Header.Get("fromnodenumber"))
			msg := taskMessage{}
			err := json.NewDecoder(r.Body).Decode(&msg)
			if err != nil {
				panic(err)
			}

			jsonMsg, err1 := json.Marshal(msg)
			if err1 != nil {
				panic(err1)
			}
			buff := bytes.NewBuffer([]byte(jsonMsg))
			client := &http.Client{}

			thisNode.TaskWord = msg.TaskWord
			thisNode.TaskBody = msg.TaskBody

			go func() {
				remapNodes()
				runTask()
			}()

			statusLog += "\ntask log: Catched task Task word: " + thisNode.TaskWord + "; Task body: " + thisNode.TaskBody + ""

			rightNode := findNearbyNode(thisNodeID, "right")
			leftNode := findNearbyNode(thisNodeID, "left")

			if fromNode < thisNodeID && rightNode != -1 {
				//refactoring more!
				count := 0
				getNodeParams(rightNode, "taskstatus")
				for nodeList[rightNode].TaskStatus != "empty" && count <= thisNodeID {
					rightNode = findNearbyNode(thisNodeID, "right")
					getNodeParams(rightNode, "taskstatus")
					statusLog += "+l" + nodeList[rightNode].TaskStatus + "l+" + strconv.Itoa(rightNode)
					count++
					if nodeList[rightNode].TaskStatus == "loaded" || nodeList[rightNode].TaskStatus == "running" {
						rightNode = -1
						break
					}
				}

				req, err := http.NewRequest("POST", "http://goappnode"+strconv.Itoa(rightNode)+".herokuapp.com/task", buff)
				if err != nil {
					panic(err)
				}
				req.Header.Set("reqtype", "taskcontent")
				req.Header.Set("fromnodenumber", strconv.Itoa(thisNodeID))
				_, err1 := client.Do(req)
				if err1 != nil {
					panic(err1)
				}
				statusLog += "\nmsg. log: Task sended to Node#" + strconv.Itoa(rightNode)
			} else if fromNode > thisNodeID && leftNode != -1 {
				//refactoring more!
				count := 0
				getNodeParams(leftNode, "taskstatus")
				for nodeList[leftNode].TaskStatus != "empty" && count <= thisNodeID {
					leftNode = findNearbyNode(thisNodeID, "left")
					getNodeParams(leftNode, "taskstatus")
					statusLog += "+l" + nodeList[leftNode].TaskStatus + "l+" + strconv.Itoa(leftNode)
					count++
					if nodeList[leftNode].TaskStatus == "loaded" || nodeList[leftNode].TaskStatus == "running" {
						leftNode = -1
						break
					}
				}

				req, err := http.NewRequest("POST", "http://goappnode"+strconv.Itoa(leftNode)+".herokuapp.com/task", buff)
				if err != nil {
					panic(err)
				}
				req.Header.Set("reqtype", "taskcontent")
				req.Header.Set("fromnodenumber", strconv.Itoa(thisNodeID))
				_, err1 := client.Do(req)
				if err1 != nil {
					panic(err1)
				}
				statusLog += "Task sended to Node#" + strconv.Itoa(leftNode)
			}
			fmt.Fprint(w, "taskcontent")
		case "taskpartresult":
			fromNode, _ := strconv.Atoi(r.Header.Get("fromnodenumber"))
			msg := resultMessage{}
			err := json.NewDecoder(r.Body).Decode(&msg)
			if err != nil {
				panic(err)
			}

			if nodeList[fromNode].TaskStatus == "completed" {
				statusLog += "\nmsg. log: Already have result from Node#" + strconv.Itoa(fromNode)
			} else {
				nodeList[fromNode].TaskResult = msg.TaskResult
				nodeList[fromNode].TaskStatus = "completed"

				tmpRes, _ := strconv.Atoi(msg.TaskResult)
				globalResult += tmpRes

				statusLog += "\nmsg. log: Received result from Node#" + strconv.Itoa(fromNode) + ",  result = " + msg.TaskResult
				statusLog += "\nres. log: Global result now = " + strconv.Itoa(globalResult)
			}

		default:
			if thisNode.TaskStatus == "empty" || thisNode.TaskStatus == "completed" {
				r.ParseForm()
				thisNode.TaskWord = r.FormValue("taskWord")
				thisNode.TaskBody = r.FormValue("taskBody")
				statusLog += "\ntask log: Get new task Task word: " + thisNode.TaskWord + "; task body: " + thisNode.TaskBody
				go sendTask()
				thisNode.TaskStatus = "loaded"
				go func() {
					remapNodes()
					runTask()
				}()

			} else {
				fmt.Fprint(w, "Can`t get your task, this node is busy")
			}
		}
	default:
		fmt.Fprint(w, "GFY")
	}

}

func logServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, statusLog)
	fmt.Fprintln(w, nodeList)
	fmt.Fprint(w, "Nodes online: ")
	fmt.Fprintln(w, activeNodes)
	fmt.Fprintln(w, "Global Result: "+strconv.Itoa(globalResult))
	go func() {
		time.Sleep(5 * time.Second)
		http.Redirect(w, r, "http://goappnode"+strconv.Itoa(thisNodeID)+".herokuapp.com/log", http.StatusSeeOther)
	}()
}

func statusServer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		jsonNodeStatus, err := json.Marshal(thisNode)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
		fmt.Fprintf(w, string(jsonNodeStatus))
	case "POST":
		fmt.Fprintln(w, "POST")
	default:
		fmt.Fprintln(w, "GFY")
	}
}

func startDiscover(w http.ResponseWriter, r *http.Request) {
	go discoverNodes()
	http.Redirect(w, r, "http://goappnode"+strconv.Itoa(thisNodeID)+".herokuapp.com", http.StatusSeeOther)

}

func cleanLogs(w http.ResponseWriter, r *http.Request) {
	statusLog = "Log "
	http.Redirect(w, r, "http://goappnode"+strconv.Itoa(thisNodeID)+".herokuapp.com", http.StatusSeeOther)
}

//------------------------------------------------------//

//prod funcs--------------------------------------------//
func initTheme() {
	themePack = map[string]string{
		"Red":    "indexrd.html",
		"Pink":   "indexpn.html",
		"Purple": "indexpr.html",
		"Indigo": "indexin.html",
		"Teal":   "indextl.html",
		"Green":  "indexgr.html",
	}
}

func periodicCheckTest(nodeID int) {
	for i := 0; i < 6; i++ {
		resp, err := http.Get("http://goappnode" + strconv.Itoa(nodeID) + ".appspot.com/status")
		if err != nil {
			statusLog += "Error: " + err.Error()
		}

		if resp.StatusCode != 200 {
			statusLog += "\nHA log: Node #" + strconv.Itoa(nodeID) + " - offline "
		} else {
			statusLog += "\nHA log: Node #" + strconv.Itoa(nodeID) + " - online "
		}

		time.Sleep(5000 * time.Millisecond)
	}
}

func sendUserMessage(nodeID int, message string) {
	msg := echoMessage{
		"Message is",
		"",
	}
	msg.Content = message
	jsonMessage, err2 := json.Marshal(msg)
	if err2 != nil {
		panic(err2)
	}
	jsonStr := []byte(jsonMessage)
	buf := bytes.NewBuffer(jsonStr)
	req, err := http.NewRequest("POST", "http://goappnode"+strconv.Itoa(nodeID)+".appspot.com/status", buf)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		statusLog = err.Error()
		fmt.Println(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	statusLog = "\nmsg. log: Response from node - " + string(respBody) + " messagee was - " + msg.Content
}

func discoverNodes() {
	activeNodes[thisNodeID] = true
	activeNodesCount = 1
	for i := 0; i < maxNodes; i++ {
		if i != thisNodeID {
			resp, err := http.Get("http://goappnode" + strconv.Itoa(i) + ".herokuapp.com/status")
			if err != nil {
				panic(err)
			}
			switch resp.StatusCode {
			case 200:
				activeNodes[i] = true
				activeNodesCount++
			default:
				activeNodes[i] = false
			}
		}
	}
	thisNode.NodeCount = activeNodesCount
}

func updateNodeStatus(nodeID int) {
	nodeList[thisNodeID] = thisNode
	if activeNodes[nodeID] == true && nodeID != thisNodeID {
		resp, err := http.Get("http://goappnode" + strconv.Itoa(nodeID) + ".herokuapp.com/status")
		if err != nil {
			panic(err)
		}
		err1 := json.NewDecoder(resp.Body).Decode(&nodeList[nodeID])
		if err1 != nil {
			panic(err1)
		}
	} else {
		nodeList[nodeID].TaskStatus = "empty"
		nodeList[nodeID].TaskResult = "empty"
	}
}

func updateNodesStatuses() {
	discoverNodes()
	for i := 0; i < 10; i++ {
		updateNodeStatus(i)
	}
}

func sendTask() {

	switch thisNodeID {
	case 0:
		/**/
		//refactoring! NEW
		rightNode := findNearbyNode(thisNodeID, "right")

		count := 0
		getNodeParams(rightNode, "taskstatus")
		for nodeList[rightNode].TaskStatus != "empty" && count <= thisNodeID {
			rightNode = findNearbyNode(thisNodeID, "right")
			getNodeParams(rightNode, "taskstatus")
			statusLog += "+r" + nodeList[rightNode].TaskStatus + "r+" + strconv.Itoa(rightNode)
			count++
			if nodeList[rightNode].TaskStatus == "loaded" {
				rightNode = -1
				break
			}
		}

		//create msg func
		msg := taskMessage{
			thisNode.TaskWord,
			thisNode.TaskBody,
		}
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		buff := bytes.NewBuffer([]byte(jsonMsg))
		client := &http.Client{}
		//

		req, err := http.NewRequest("POST", "http://goappnode"+strconv.Itoa(rightNode)+".herokuapp.com/task", buff)
		if err != nil {
			panic(err)
		}
		req.Header.Set("reqtype", "taskcontent")
		req.Header.Set("fromnodenumber", strconv.Itoa(thisNodeID))
		_, err1 := client.Do(req)
		if err1 != nil {
			panic(err1)
		}
		statusLog += "\nmsg. log: Task sended to Node#" + strconv.Itoa(thisNodeID+1)
	case 9:
		/**/
		//refactoring! NEW
		leftNode := findNearbyNode(thisNodeID, "left")

		count := 0
		getNodeParams(leftNode, "taskstatus")
		for nodeList[leftNode].TaskStatus != "empty" && count <= thisNodeID {
			leftNode = findNearbyNode(thisNodeID, "left")
			getNodeParams(leftNode, "taskstatus")
			statusLog += "+l" + nodeList[leftNode].TaskStatus + "l+" + strconv.Itoa(leftNode)
			count++
			if nodeList[leftNode].TaskStatus == "loaded" {
				leftNode = -1
				break
			}
		}

		//create msg func
		msg := taskMessage{
			thisNode.TaskWord,
			thisNode.TaskBody,
		}
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		buff := bytes.NewBuffer([]byte(jsonMsg))
		client := &http.Client{}
		//

		req, err := http.NewRequest("POST", "http://goappnode"+strconv.Itoa(leftNode)+".herokuapp.com/task", buff)
		if err != nil {
			panic(err)
		}
		req.Header.Set("reqtype", "taskcontent")
		req.Header.Set("fromnodenumber", strconv.Itoa(thisNodeID))
		_, err1 := client.Do(req)
		if err1 != nil {
			panic(err1)
		}
		statusLog += "\nmsg. log: Task sended to Node#" + strconv.Itoa(thisNodeID-1)
	default:

		leftNode := findNearbyNode(thisNodeID, "left")
		rightNode := findNearbyNode(thisNodeID, "right")
		/*
			statusLog += "rightNode" + strconv.Itoa(rightNode)
			statusLog += "leftNode" + strconv.Itoa(leftNode)
		*/

		if leftNode != -1 {
			count := 0
			getNodeParams(leftNode, "taskstatus")
			for nodeList[leftNode].TaskStatus != "empty" && count <= thisNodeID {
				leftNode = findNearbyNode(thisNodeID, "left")
				getNodeParams(leftNode, "taskstatus")
				statusLog += "+l" + nodeList[leftNode].TaskStatus + "l+" + strconv.Itoa(leftNode)
				count++
				if nodeList[leftNode].TaskStatus == "loaded" || nodeList[leftNode].TaskStatus == "running" {
					leftNode = -1
					break
				}
			}

			//create msg func
			msg := taskMessage{
				thisNode.TaskWord,
				thisNode.TaskBody,
			}
			jsonMsg, err := json.Marshal(msg)
			if err != nil {
				panic(err)
			}
			buff := bytes.NewBuffer([]byte(jsonMsg))
			client := &http.Client{}

			req, err := http.NewRequest("POST", "http://goappnode"+strconv.Itoa(leftNode)+".herokuapp.com/task", buff)
			if err != nil {
				panic(err)
			}
			req.Header.Set("reqtype", "taskcontent")
			req.Header.Set("fromnodenumber", strconv.Itoa(thisNodeID))
			_, err1 := client.Do(req)
			if err1 != nil {
				panic(err1)
			}
			statusLog += "\nmsg. log: Task sended to Node#" + strconv.Itoa(leftNode)
		}

		if rightNode != -1 {
			count := 0
			getNodeParams(rightNode, "taskstatus")
			for nodeList[rightNode].TaskStatus != "empty" && count <= thisNodeID {
				rightNode = findNearbyNode(thisNodeID, "right")
				getNodeParams(rightNode, "taskstatus")
				statusLog += "+r" + nodeList[rightNode].TaskStatus + "r+" + strconv.Itoa(rightNode)
				count++
				if nodeList[rightNode].TaskStatus == "loaded" || nodeList[rightNode].TaskStatus == "running" {
					rightNode = -1
					break
				}
			}

			//create msg func
			msg := taskMessage{
				thisNode.TaskWord,
				thisNode.TaskBody,
			}
			jsonMsg, err := json.Marshal(msg)
			if err != nil {
				panic(err)
			}
			buff := bytes.NewBuffer([]byte(jsonMsg))
			client := &http.Client{}
			//

			req, err := http.NewRequest("POST", "http://goappnode"+strconv.Itoa(rightNode)+".herokuapp.com/task", buff)
			if err != nil {
				panic(err)
			}
			req.Header.Set("reqtype", "taskcontent")
			req.Header.Set("fromnodenumber", strconv.Itoa(thisNodeID))
			_, err1 := client.Do(req)
			if err1 != nil {
				panic(err1)
			}
			statusLog += "\nmsg. log: Task sended to Node#" + strconv.Itoa(rightNode)
		}
	}
	//go updateNodesStatuses()
}

func findNearbyNode(startId int, direction string) int {
	switch direction {
	case "right":
		rightNode := -1
		for i := startId; i < 10; i++ {
			if i != startId && activeNodes[i] {
				rightNode = i
				return rightNode
			}
		}
	case "left":
		leftNode := -1
		for i := startId; i > -1; i-- {
			if i != startId && activeNodes[i] {
				leftNode = i
				return leftNode
			}
		}
	}
	return -1
}

func getNodeParams(nodeID int, header string /*, syncer *sync.WaitGroup*/) {
	//defer syncer.Done()
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://goappnode"+strconv.Itoa(nodeID)+".herokuapp.com/task", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("getnodeparams", header)
	resp, err1 := client.Do(req)
	if err1 != nil {
		panic(err1)
	}
	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		panic(err2)
	}

	switch header {
	case "taskstatus":
		nodeList[nodeID].TaskStatus = string(bodyBytes)
	case "taskresult":
		nodeList[nodeID].TaskResult = string(bodyBytes)
	case "taskword":
		nodeList[nodeID].TaskWord = string(bodyBytes)
	case "taskfragmentbody":
		nodeList[nodeID].TaskFragmentBody = string(bodyBytes)
	case "tasktody":
		nodeList[nodeID].TaskBody = string(bodyBytes)
	case "addtask":
		hastask := string(bodyBytes)
		nodeList[nodeID].HasAddTask, _ = strconv.Atoi(hastask)
	}
}

func remap(nodeID int, tempID int) {
	nodesMap[nodeID] = tempID
}

func remapNodes() {
	updateNodesStatuses()
	nodesMap = make(map[int]int)
	nodesTaskGroup = make([]int, activeNodesCount)
	tempID := 0
	for i := 0; i < maxNodes; i++ {
		if activeNodes[i] {
			nodesMap[i] = tempID
			nodesTaskGroup[tempID] = i
			statusLog += "\ntask log: Task mapping. Node#" + strconv.Itoa(i) + " get new subID = " + strconv.Itoa(tempID)
			tempID++
		}
	}
}

func runTask() {
	thisNode.TaskStatus = "running"
	thisNode.HasAddTask = 999
	globalResult = 0
	go HAprocess()
	if activeNodesCount != maxNodes {
		fragmentLen := (len(thisNode.TaskBody) / activeNodesCount)
		taskFragmentLen = fragmentLen
		//statusLog += "fragmentLen=" + strconv.Itoa(fragmentLen)
		if nodesMap[thisNodeID] != activeNodesCount-1 {
			thisNode.TaskFragmentBody = thisNode.TaskBody[nodesMap[thisNodeID]*fragmentLen : nodesMap[thisNodeID]*fragmentLen+fragmentLen]

			//go allResultProcess()
			res := strings.Count(thisNode.TaskFragmentBody, thisNode.TaskWord)
			thisNode.TaskResult = strconv.Itoa(res)
			globalResult += res
			time.Sleep(40 * time.Second)
			thisNode.TaskStatus = "completed"
			go resetTaskStatus()

			for i := 0; i < len(nodesTaskGroup); i++ {
				if nodesTaskGroup[i] != thisNodeID {
					//send result
					msg := resultMessage{
						thisNode.TaskResult,
					}
					jsonMsg, err := json.Marshal(msg)
					if err != nil {
						panic(err)
					}
					buff := bytes.NewBuffer([]byte(jsonMsg))
					client := &http.Client{}

					req, err := http.NewRequest("POST", "http://goappnode"+strconv.Itoa(nodesTaskGroup[i])+".herokuapp.com/task", buff)
					if err != nil {
						panic(err)
					}
					req.Header.Set("reqtype", "taskpartresult")
					req.Header.Set("fromnodenumber", strconv.Itoa(thisNodeID))
					_, err1 := client.Do(req)
					if err1 != nil {
						panic(err1)
					}
					statusLog += "\nmsg. log: Result sended to Node#" + strconv.Itoa(nodesTaskGroup[i])
				}
			}
		} else {
			thisNode.TaskFragmentBody = thisNode.TaskBody[nodesMap[thisNodeID]*fragmentLen : len(thisNode.TaskBody)]
			go HAprocess()
			//go allResultProcess()
			res := strings.Count(thisNode.TaskFragmentBody, thisNode.TaskWord)
			thisNode.TaskResult = strconv.Itoa(res)
			globalResult += res
			time.Sleep(20 * time.Second)

			thisNode.TaskStatus = "completed"

			for i := 0; i < len(nodesTaskGroup); i++ {
				if nodesTaskGroup[i] != thisNodeID {
					//send result
					msg := resultMessage{
						thisNode.TaskResult,
					}
					jsonMsg, err := json.Marshal(msg)
					if err != nil {
						panic(err)
					}
					buff := bytes.NewBuffer([]byte(jsonMsg))
					client := &http.Client{}

					req, err := http.NewRequest("POST", "http://goappnode"+strconv.Itoa(nodesTaskGroup[i])+".herokuapp.com/task", buff)
					if err != nil {
						panic(err)
					}
					req.Header.Set("reqtype", "taskpartresult")
					req.Header.Set("fromnodenumber", strconv.Itoa(thisNodeID))
					_, err1 := client.Do(req)
					if err1 != nil {
						panic(err1)
					}
					statusLog += "\nmsg. log: Result sended to Node#" + strconv.Itoa(nodesTaskGroup[i])
				}
			}
		}
	}
}

func HAprocess() {
	for {
		time.Sleep(5 * time.Second)

		discoverNodes()

		runAddTask := false
		fallenNode := 999
		for i := 0; i < len(nodesTaskGroup); i++ {
			if !activeNodes[nodesTaskGroup[i]] {
				fallenNode = nodesTaskGroup[i]
				statusLog += "\nHA log: Node #" + strconv.Itoa(fallenNode) + " is down!"
				if nodeList[fallenNode].TaskStatus != "completed" {
					statusLog += "\nHA log: Down Node #" + strconv.Itoa(fallenNode) + " is not finish the task!"
					for n := 0; n < activeNodesCount; n++ {
						if n != i {
							getNodeParams(nodesTaskGroup[n], "addtask")
							statusLog += "\nHA log: Node #" + strconv.Itoa(n) + " has add task = " + strconv.Itoa(nodeList[nodesTaskGroup[n]].HasAddTask)
							if nodeList[n].HasAddTask == i {
								runAddTask = false
							} else {
								runAddTask = true
								statusLog += "\nHA log: Need to start node task"
							}
						}
					}
				} else {
					runAddTask = false
				}
				if runAddTask {
					thisNode.HasAddTask = fallenNode
					statusLog += "\nHA log: HA status changed"
					statusLog += "\nHA log: Starting down node task"
					thisNode.TaskStatus = "running"
					fragmentLen := (len(thisNode.TaskBody) / len(nodesTaskGroup))
					statusLog += "\nHA log: Down node fragmentLen=" + strconv.Itoa(fragmentLen)
					tmpRes := 0
					if nodesMap[fallenNode] != len(nodesTaskGroup)-1 {
						nodeList[fallenNode].TaskFragmentBody = thisNode.TaskBody[nodesMap[fallenNode]*fragmentLen : nodesMap[fallenNode]*fragmentLen+fragmentLen]
						nodeList[fallenNode].TaskResult = strconv.Itoa(strings.Count(nodeList[fallenNode].TaskFragmentBody, thisNode.TaskWord))
						thisNode.TaskStatus = "completed"
						nodeList[fallenNode].TaskStatus = "completed"
						statusLog += "\nHA log: Down node task result (not end node) = " + nodeList[nodesTaskGroup[i]].TaskResult
						tmpRes, _ = strconv.Atoi(nodeList[fallenNode].TaskResult)
					} else {
						nodeList[fallenNode].TaskFragmentBody = thisNode.TaskBody[nodesMap[fallenNode]*fragmentLen : len(thisNode.TaskBody)]
						nodeList[fallenNode].TaskResult = strconv.Itoa(strings.Count(nodeList[fallenNode].TaskFragmentBody, thisNode.TaskWord))
						thisNode.TaskStatus = "completed"
						nodeList[fallenNode].TaskStatus = "completed"
						statusLog += "\nHA log: Down node task result (end node) = " + nodeList[nodesTaskGroup[i]].TaskResult
						tmpRes, _ = strconv.Atoi(nodeList[fallenNode].TaskResult)
					}
					globalResult += tmpRes
					statusLog += "\nHA log: Down node task result added to global result"

					msg := resultMessage{
						nodeList[fallenNode].TaskResult,
					}
					jsonMsg, err := json.Marshal(msg)
					if err != nil {
						panic(err)
					}
					buff := bytes.NewBuffer([]byte(jsonMsg))
					client := &http.Client{}

					for i := 0; i < len(nodesTaskGroup); i++ {
						statusLog += "\nmsg. log: Compare nodes, Node in task group - " + strconv.Itoa(nodesTaskGroup[i]) + ", node in map - " + strconv.Itoa(nodesMap[nodesTaskGroup[i]])
						if nodesTaskGroup[i] != thisNodeID {
							statusLog += "\nmsg. log: Compare nodes, Node in task group - " + strconv.Itoa(nodesTaskGroup[i]) + ", node in map - " + strconv.Itoa(nodesMap[nodesTaskGroup[i]])
							req, err := http.NewRequest("POST", "http://goappnode"+strconv.Itoa(nodesTaskGroup[i])+".herokuapp.com/task", buff)
							if err != nil {
								panic(err)
							}
							req.Header.Set("reqtype", "taskpartresult")
							req.Header.Set("fromnodenumber", strconv.Itoa(fallenNode))
							_, err1 := client.Do(req)
							if err1 != nil {
								panic(err1)
							}
							statusLog += "\nmsg. log: Result sended to Node#" + strconv.Itoa(nodesTaskGroup[i])
						}
					}
				}
			}
		}
	}
}

func allResultProcess() {
	time.Sleep(7 * time.Second)

	ready := true
	for i := 0; i < len(nodesTaskGroup); i++ {
		if nodeList[nodesTaskGroup[i]].TaskStatus != "completed" {
			ready = false
		}
	}
	if ready {
		/*
			allResult := 0

			for i := 0; i < len(nodesTaskGroup); i++ {
				tmp, _ := strconv.Atoi(nodeList[nodesTaskGroup[i]].TaskResult)
				allResult += tmp
			}
		*/
		statusLog += "\nres. log: FINAL RESULT = " + strconv.Itoa(globalResult)
	}
}

func resetTaskStatus() {
	time.Sleep(5 * time.Second)
	thisNode.TaskStatus = "empty"
}

//-----------------------------------------------------//
