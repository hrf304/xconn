package tools

import "sync"

type Notifier struct {
	dataChans chan map[string]interface{}
	subs     map[string][]func(interface{})
	subMutex sync.Mutex
}

func NewNotifier(dataSize int)*Notifier{
	n := &Notifier{
		dataChans:make(chan map[string]interface{}, dataSize),
		subs: make(map[string][]func(interface{})),
	}

	return n
}

func (n *Notifier)Start(){
	n.startPublishProcess()
}

func (n *Notifier)Subscribe(topic string, handler func(interface{})){
	n.subMutex.Lock()
	defer n.subMutex.Unlock()

	if _, ok := n.subs[topic]; ok{
		n.subs[topic] = append(n.subs[topic], handler)
	}else{
		n.subs[topic] = []func(interface{}){handler}
	}
}

func (n *Notifier)Publish(topic string, data interface{}){
	n.dataChans <- map[string]interface{}{"topic": topic, "data": data}
}

func (n *Notifier)startPublishProcess(){
	go func() {
		for{
			select {
			case item := <- n.dataChans:
				topic := item["topic"].(string)
				data := item["data"]

				n.subMutex.Lock()
				if handlers, ok := n.subs[topic]; ok{
					for i := range handlers{
						go handlers[i](data)
					}
				}
				n.subMutex.Unlock()
			}
		}
	}()
}
