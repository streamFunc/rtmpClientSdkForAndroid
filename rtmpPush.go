package rtmpSdk

import (
	"fmt"
	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-flv"
	"github.com/yapingcat/gomedia/go-rtmp"
	"net"
	"net/url"
	"os"
	"time"
)

var globalClient RtmpClient

type RtmpClient struct {
	rtmpUrl string
	isReady bool
	isStop  bool
	pts     uint32
	c       net.Conn
	client  *rtmp.RtmpClient
}

//var rtmpUrl = flag.String("url", "rtmp://127.0.0.1/live/test", "publish rtmp url")
//var flvFile = flag.String("flv", "test.flv", "push flv file to server")

func publish(fileName string, cli *rtmp.RtmpClient) {
	fmt.Println(fileName)
	f := flv.CreateFlvReader()
	f.OnFrame = func(cid codec.CodecID, frame []byte, pts, dts uint32) {
		if cid == codec.CODECID_VIDEO_H264 {
			cli.WriteVideo(cid, frame, pts, dts)
			time.Sleep(time.Millisecond * 20)
		} else if cid == codec.CODECID_VIDEO_H265 {
			cli.WriteVideo(cid, frame, pts, dts)
			time.Sleep(time.Millisecond * 20)
		} else if cid == codec.CODECID_AUDIO_AAC {
			cli.WriteAudio(cid, frame, pts, dts)
		}
	}
	fd, _ := os.Open(fileName)
	defer fd.Close()
	cache := make([]byte, 4096)
	for {
		n, err := fd.Read(cache)
		if err != nil {
			fmt.Println(err)
			break
		}
		f.Input(cache[0:n])
	}
}



func StartConnect(rtmpUrl string) {
	globalClient = RtmpClient{}
	globalClient.isReady = false
	globalClient.pts = 0

	u, err := url.Parse(rtmpUrl)
	if err != nil {
		panic(err)
	}
	host := u.Host
	if u.Port() == "" {
		host += ":1935"
	}
	//connect to remote rtmp server
	globalClient.c, err = net.Dial("tcp4", host)
	if err != nil {
		fmt.Println("connect failed", err)
		return
	}
	fmt.Println("connect success...")

	isReady := make(chan struct{})
	//创建rtmp client 使能复杂握手,rtmp推流
	globalClient.client = rtmp.NewRtmpClient(rtmp.WithComplexHandshake(), rtmp.WithEnablePublish())

	//监听状态变化,STATE_RTMP_PUBLISH_START 状态通知推流
	globalClient.client.OnStateChange(func(newState rtmp.RtmpState) {
		if newState == rtmp.STATE_RTMP_PUBLISH_START {
			fmt.Println("lby test ready for publish")
			globalClient.isReady = true
			close(isReady)
		}
	})

	globalClient.client.SetOutput(func(data []byte) error {
		_, err := globalClient.c.Write(data)
		return err
	})

	globalClient.client.Start(rtmpUrl)
	fmt.Println("client Start...")

	go func() {
		buf := make([]byte, 4096)
		n := 0
		for {
			if globalClient.isStop {
				return
			}
			n, err = globalClient.c.Read(buf)
			if err != nil {
				continue
			}
			globalClient.client.Input(buf[:n])
		}
	}()
}

func StartPush(nal []byte) {
	if globalClient.isReady {
		globalClient.client.WriteVideo(codec.CODECID_VIDEO_H264, nal, globalClient.pts, globalClient.pts)
		globalClient.pts += 3600
	}
}

func StopConnect() {
	globalClient.isStop = true
	globalClient.isReady = false
}


/*func main() {
	fmt.Println("start...")
	StartConnect("rtmp://47.92.86.188/live/test")
	fd, _ := os.Open("test.flv")
	defer fd.Close()
	cache := make([]byte, 4096)
	fmt.Println("start 111...")
	for {
		n, err := fd.Read(cache)
		if err != nil {
			fmt.Println(err)
			break
		}
		StartPush(cache[0:n])
	}
	StopConnect()

	return
	flag.Parse()
	u, err := url.Parse(*rtmpUrl)
	if err != nil {
		panic(err)
	}
	host := u.Host
	if u.Port() == "" {
		host += ":1935"
	}
	//connect to remote rtmp server
	c, err := net.Dial("tcp4", host)
	if err != nil {
		fmt.Println("connect failed", err)
		return
	}

	isReady := make(chan struct{})

	//创建rtmp client 使能复杂握手,rtmp推流
	cli := rtmp.NewRtmpClient(rtmp.WithComplexHandshake(), rtmp.WithEnablePublish())

	//监听状态变化,STATE_RTMP_PUBLISH_START 状态通知推流
	cli.OnStateChange(func(newState rtmp.RtmpState) {
		if newState == rtmp.STATE_RTMP_PUBLISH_START {
			fmt.Println("ready for publish")
			close(isReady)
		}
	})

	cli.SetOutput(func(data []byte) error {
		_, err := c.Write(data)
		return err
	})

	go func() {
		<-isReady
		fmt.Println("start to read flv")
		publish(*flvFile, cli)
	}()

	cli.Start(*rtmpUrl)
	buf := make([]byte, 4096)
	n := 0
	for err == nil {
		n, err = c.Read(buf)
		if err != nil {
			continue
		}
		cli.Input(buf[:n])
	}
	fmt.Println(err)
}*/
