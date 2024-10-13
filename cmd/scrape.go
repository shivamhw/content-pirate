package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"log"

	"github.com/shivamhw/reddit-pirate/store"
	"github.com/spf13/cobra"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type scrapper struct {
	DstStore   store.Store
	AuthCfg    AuthCfg
	Subreddits []string
	reddit     *reddit.Client
	sCfg       *scrapeCfg
	ctx        context.Context
	dstPath    DstPath
}

type DstPath struct {
	ImgPath  string
	VidPath  string
	BasePath string
}

type AuthCfg struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type scrapeCfg struct {
	dstDir       string
	authCfg      string
	subreddits   string
	postId       string
	duration     string
	skipVideo    bool
	cleanOnStart bool
	combineDir   bool
	imgWorker	 int
	vidWorker	 int
	redWorker    int
}

type Mediums struct {
	subq      chan string
	postq     chan Post
	imgq      chan Job
	vidq      chan Job
	swg       sync.WaitGroup
	post_done chan bool
}

type Job struct {
	src  string
	dst  string
	name string
}

type Post struct {
	media     string
	link      string
	title     string
	id        string
	subreddit string
	ext       string
}

const (
	IMGS = "imgs"
	VIDS = "vids"
)

var (
	IMG_SUFFIX = []string{".jpg", ".jpeg", ".png", ".gif"}
	VID_SUFFIX = []string{".mp4"}
	sCfg       scrapeCfg
	aCfg       AuthCfg
	imgCounter int64
	vidCounter int64
)

var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Long:  "Scrapes subreddit for videos and imgs",
	Short: "scrapes subreddit",
	Run:   scrapperHandler,
}

func init() {
	scrapeCmd.Flags().StringVar(&sCfg.dstDir, "dir", "./download", "dst folder for downloads")
	scrapeCmd.Flags().StringVar(&sCfg.subreddits, "subs", "./subreddits.json", "list of subreddits")
	scrapeCmd.Flags().StringVar(&sCfg.authCfg, "auth", "./reddit.json", "auth config for reddit")
	scrapeCmd.Flags().StringVar(&sCfg.postId, "post-id", "", "post id")
	scrapeCmd.Flags().StringVar(&sCfg.duration, "duration", "day", "duration")
	scrapeCmd.Flags().BoolVar(&sCfg.skipVideo, "skip-vid", true, "skip video download")
	scrapeCmd.Flags().BoolVar(&sCfg.combineDir, "combine", true, "combine folders")
	scrapeCmd.Flags().BoolVar(&sCfg.cleanOnStart, "cleanOnStart", true, "clean folders")
	scrapeCmd.Flags().IntVar(&sCfg.imgWorker, "img-worker", 10, "nof img proccesing worker")
	scrapeCmd.Flags().IntVar(&sCfg.vidWorker, "vid-worker", 5, "nof vid proccesing worker")
	scrapeCmd.Flags().IntVar(&sCfg.redWorker, "reddit-worker", 10, "nof reddit proccesing worker")
}

func getCfgFromJson(filePath string, v interface{}) {
	file, _ := os.Open(filePath)
	defer file.Close()

	data, _ := io.ReadAll(file)
	if err := json.Unmarshal(data, v); err != nil {
		log.Fatal("fat gta")
	}
}

func (s scrapper) createStructure() {
	if sCfg.cleanOnStart {
		err := os.RemoveAll(s.dstPath.getBasePath())
		if err != nil {
			log.Print("err while deleting dir structure ", err)
		}else{
			log.Print("cleanup success")
		}
	}
	for _, f := range s.Subreddits {
		log.Println("creating ", s.dstPath.getImgPath(f))
		log.Println("creating ", s.dstPath.getVidPath(f))
		s.DstStore.CreateDir(s.dstPath.getImgPath(f))
		s.DstStore.CreateDir(s.dstPath.getVidPath(f))
	}
}

func scrapperHandler(cmd *cobra.Command, args []string) {
	scr := scrapper{
		sCfg:   &sCfg,
		ctx:    context.Background(),
		dstPath: DstPath{
			BasePath: sCfg.dstDir,
			ImgPath:  IMGS,
			VidPath:  VIDS,
		},
	}
	// load auth1
	getCfgFromJson(sCfg.authCfg, &scr.AuthCfg)
	// load sub reddit
	getCfgFromJson(sCfg.subreddits, &scr.Subreddits)
	// create auth
	credentials := reddit.Credentials{ID: scr.AuthCfg.ID, Secret: scr.AuthCfg.Secret, Username: scr.AuthCfg.Username, Password: scr.AuthCfg.Password}
	c, err := reddit.NewClient(credentials)
	if err != nil {
		log.Printf("err creating client %s, using default client", err)
		c = reddit.DefaultClient()
	}
	scr.reddit = c
	//creating dir struct
	scr.DstStore = store.FileStore{Dir: sCfg.dstDir}
	scr.createStructure()
	scr.Run()
	log.Printf("Processed Imgs : %d", imgCounter)
	log.Printf("Processed vids : %d", vidCounter)
}

func isImgLink(link string) bool {
	for _, suff := range IMG_SUFFIX {
		if strings.HasSuffix(link, suff) {
			return true
		}
	}
	return false
}

func parsePost(subreddit string, p *reddit.Post) (*[]Post, error) {
	//do we need to parse in first place?
	var final_posts []Post
	if strings.Contains(p.URL, "/gallery/") {
		log.Print("found gallery ", p.URL)
		for _, item := range p.GalleryData.Items {
			link := fmt.Sprintf("https://i.redd.it/%s.%s", item.MediaID, getMIME(p.MediaMetadata[item.MediaID]))
			log.Printf("created link %s for gal %s %s", link, p.Title, item.MediaID)
			if isImgLink(link) {
				post := Post{
					id:        p.ID,
					title:     fmt.Sprintf("%s_GAL_%s", p.Title, item.MediaID[:len(item.MediaID)-3]),
					media:     IMGS,
					ext:       getMIME(p.MediaMetadata[item.MediaID]),
					link:      link,
					subreddit: subreddit,
				}
				final_posts = append(final_posts, post)
			}
		}
		return &final_posts, nil
	}

	if isImgLink(p.URL) {
		post := Post{
			id:        p.ID,
			title:     p.Title,
			media:     IMGS,
			ext:       strings.Split(p.URL, ".")[len(strings.Split(p.URL, "."))-1],
			link:      p.URL,
			subreddit: subreddit,
		}
		final_posts = append(final_posts, post)
		return &final_posts, nil
	}
	if !sCfg.skipVideo && p.Media.RedditVideo.FallbackURL != "" {
		post := Post{
			id:        p.ID,
			title:     p.Title,
			media:     VIDS,
			link:      p.Media.RedditVideo.FallbackURL,
			ext:       "mp4",
			subreddit: subreddit,
		}
		final_posts = append(final_posts, post)
		return &final_posts, nil
	}
	return nil, fmt.Errorf("can not parse %s this postURL %s", p.ID, p.URL)

}

func (s scrapper) EmitPosts(subreddit string) ([]Post, error) {
	var final_posts []Post
	nextToken := ""
	for {
		posts, resp, err := s.reddit.Subreddit.TopPosts(s.ctx, subreddit, &reddit.ListPostOptions{
			ListOptions: reddit.ListOptions{
				Limit: 100,
				After: nextToken,
			},
			Time: sCfg.duration,
		})
		if err != nil {
			if strings.Contains(err.Error(), "429") {
				log.Printf("HIT rate limit wait 2 sec")
				time.Sleep(2 * time.Second)
				continue
			} else {
				return nil, err
			}
		}
		// adding post
		for _, p := range posts {
			post, err := parsePost(subreddit, p)
			if err != nil {
				log.Printf("post %s: %s", p.Title, err)
				continue
			}
			final_posts = append(final_posts, *post...)
		}

		nextToken = resp.After
		if nextToken == "" {
			break
		}
	}
	return final_posts, nil
}

func (s scrapper) downloadJob(j Job) error {
	resp, err := http.Get(j.src)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s because %s code", j.src, err)
	}
	defer resp.Body.Close()

	//save to dir
	log.Printf("saving %s to filesystem ", j.name)
	err = s.DstStore.Write(filepath.Join(j.dst, j.name), resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file %s to %s as %s", j.name, j.dst, err)
	}
	return nil
}

func (s scrapper) processImg(j Job) {
	//download file
	if err := s.downloadJob(j); err != nil {
		log.Printf("failed while downloading imgs %s ", err)
	}
	atomic.AddInt64(&imgCounter, 1)
}

func (s scrapper) processVid(j Job) {
	if err := s.downloadJob(j); err != nil {
		log.Printf("failed while downloading vid %s ", err)
	}
	atomic.AddInt64(&vidCounter, 1)
}

func (s scrapper) subWorker(id int, m *Mediums, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("started sub worker %d\n", id)
	for r := range m.subq {
		posts, err := s.EmitPosts(r)
		if err != nil {
			log.Fatalf("%s", err)
		}
		for _, p := range posts {
			m.postq <- p
		}
	}
	fmt.Println("sub worker exits ", id)
}

func (s scrapper) imgWorker(id int, m *Mediums) {
	defer m.swg.Done()
	fmt.Println("starting img woker ", id)
	for j := range m.imgq {
		fmt.Println("processing img ", j.name)
		s.processImg(j)
	}
	fmt.Println("Exited img worker ", id)
}

func (s scrapper) vidWorker(id int, m *Mediums) {
	defer m.swg.Done()
	fmt.Println("starting vid woker ", id)
	for j := range m.vidq {
		fmt.Println("processing VIDEO ", j.name, j.src)
		s.processVid(j)
	}
	fmt.Println("Exited vid worker ", id)

}

func (s scrapper) startWorkers(m *Mediums) {
	var sub_wg sync.WaitGroup

	for i := 0; i < sCfg.redWorker; i++ {
		sub_wg.Add(1)
		go s.subWorker(i, m, &sub_wg)
	}

	for i := 0; i < sCfg.imgWorker; i++ {
		m.swg.Add(1)
		go s.imgWorker(i, m)
	}

	for i := 0; i < sCfg.vidWorker; i++ {
		m.swg.Add(1)
		go s.vidWorker(i, m)
	}

	sub_wg.Wait()
	close(m.postq)
}

func (s scrapper) getPostById(id string) {
	post, resp, err := s.reddit.Post.Get(s.ctx, id)
	if err != nil {
		fmt.Printf("Error getting post: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Print the parsed post data directly
	fmt.Printf("Post details: %+v\n", post)

	// If you still need to read the raw body, do so before closing the body.
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}
	fmt.Printf("Raw JSON response: %s\n", data)
}

func (s scrapper) Run() {
	if s.sCfg.postId != "" {
		log.Printf("only fetching data for %s", sCfg.postId)
		s.getPostById(sCfg.postId)
		return
	}
	var mwg sync.WaitGroup
	m := &Mediums{
		subq:  make(chan string),
		postq: make(chan Post),
		imgq:  make(chan Job, 10),
		vidq:  make(chan Job, 1000),
	}
	go s.startWorkers(m)
	mwg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer func() {
			close(m.subq)
		}()
		for _, s := range s.Subreddits {
			fmt.Println("scrapping ", s)
			m.subq <- s
		}
	}(&mwg)
LOOP:
	for {
		select {
		case v, ok := <-m.postq:
			if !ok {
				close(m.imgq)
				close(m.vidq)
				break LOOP
			}
			if v.media == VIDS {
				m.vidq <- Job{
					src:  v.link,
					dst:  s.dstPath.getVidPath(v.subreddit),
					name: fmt.Sprintf("%s_%s.%s", v.id, v.title, v.ext),
				}
			}

			if v.media == IMGS {
				m.imgq <- Job{
					src:  v.link,
					dst:  s.dstPath.getImgPath(v.subreddit),
					name: fmt.Sprintf("%s_%s.%s", v.id, v.title, v.ext),
				}
			}
		case <-m.post_done:
			break LOOP
		}
	}
	m.swg.Wait()
}

func (d DstPath) getSubredditPath(r string) string {
	return filepath.Join(d.BasePath, r)
}

func (d DstPath) getBasePath() string {
	return d.BasePath
}

func (d DstPath) getImgPath(r string) string {
	if sCfg.combineDir {
		return filepath.Join(d.BasePath, d.ImgPath)
	}
	return filepath.Join(d.BasePath, r, d.ImgPath)
}

func (d DstPath) getVidPath(r string) string {
	if sCfg.combineDir {
		return filepath.Join(d.BasePath, d.VidPath)
	}
	return filepath.Join(d.BasePath, r, d.VidPath)
}

func getMIME(media reddit.MediaData) string {
	mime := strings.Split(media.MIME, "/")
	if len(mime) == 1 {
		return "jpg"
	}
	return mime[1]
}
