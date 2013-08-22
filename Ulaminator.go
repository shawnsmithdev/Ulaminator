package main

import (
    "bufio"
    "flag"
    "fmt"
    "image"
    "image/color"
    "image/png"
    "math"
    "os"
    "runtime"
    "time"
)

const MIN_SIZE int = 3     //Prevents index out of bounds
const MAX_SIZE int = 16000 //That's a huge image.
const NUM_THREAD int = 4
const VERSION = "1.21.3"

func main() {
    //flags
    var size int    //width and height of image
    var grey bool   //if the image is monochrome or greyscale
    var name string //image filename

    flag.IntVar(&size, "s", 0, "specify image width (default prompts user)")
    flag.BoolVar(&grey, "g", true, "set false to create a monochrome image")
    flag.StringVar(&name, "o", "ulam.png", "attempt to write image at given filename")
    flag.Parse()

    //Display Info
    fmt.Println("Ulaminator", VERSION, "by Shawn Paul Smith")
    if flag.NArg() > 0 {
        flag.PrintDefaults()
        return
    }
    if grey {
        fmt.Println("Generating a greyscale Ulam Spiral")
        fmt.Println("Lighter Pixels = More Prime Factors")
    } else {
        fmt.Println("Generating a monochrome Ulam Spiral")
        fmt.Println("Black Pixels are Primes, White are Not")
    }
    fmt.Printf("Ulaminator always generates square images.\n")

    //Setup
    runtime.GOMAXPROCS(NUM_THREAD) //GO's schedular needs some work...
    if size == 0 || !isValidSize(size) {
        size = getSize()
    } //ask user for image width
    if size == 0 { //user fail
        fmt.Printf("Please Try Again!\n")
        return
    }
    pixcount := size * size
    tab := make([]uint8, pixcount+1) //We use unit8 to save ram

    //Calculate prime factor counts
    fmt.Printf("Now calculating %d prime factorizations...\n", pixcount)
    start := time.Now()
    primes := factCount(tab[:])
    dtime := time.Now().Sub(start)

    //statistics
    fmt.Printf("Calculation took %s\n", dtime.String())
    fmt.Printf("Average time per pixel: %d(ns)\n", dtime.Nanoseconds()/int64(pixcount))
    fmt.Printf("Found %v primes\n", primes)
    fmt.Printf("Ratio: %v pixels per prime\n", float64(pixcount)/float64(primes))

    //render raw image
    fmt.Println("Now rendering image...")
    start = time.Now()
    raw := renderImage(tab, size, grey)
    fmt.Printf("Rendering took %s,\n", time.Now().Sub(start).String())

    //wrte png to disk
    fmt.Println("Now compressing and writing image...")
    start = time.Now()
    err := writePng(raw, name)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Printf("Image written to %s in %s\n", name, time.Now().Sub(start).String())
    }
}

//Queries the user for the size of the ulam spiral to generate
func getSize() int {
    inbuf := bufio.NewReader(os.Stdin)
    fmt.Printf("How many pixels wide?\n:")
    var count int
    if countstrng, err := inbuf.ReadString('\n'); err != nil {
        fmt.Printf("Error:%v\n", err)
        return 0
    } else {
        if _, err2 := fmt.Sscanf(countstrng, "%d", &count); err2 != nil {
            fmt.Printf("Error:%v\n", err2)
            return 0
        }
    }

    //Bounds checking
    if isValidSize(count) {
        return count
    }
    return 0 //invalid
}

//Bounds checking
func isValidSize(size int) bool {
    if size < MIN_SIZE { //too little
        fmt.Printf("Error: %d invalid, min is %d\n", size, MIN_SIZE)
        return false
    } else if size > MAX_SIZE { //too big
        fmt.Printf("Error: %d invalid, max is %d\n", size, MAX_SIZE)
        return false
    }
    return true
}

//Generates a ulam spiral
func renderImage(tab []uint8, size int, isGreyscale bool) image.Image {
    //the most prime factors any number <= size will have
    max := int(1 + math.Log2(float64(size*size)))
    //A table of max grays
    colorTab := make([]color.NRGBA, max)
    for i := range colorTab {
        grey := uint8((i + 1) * 255 / max)
        colorTab[i] = color.NRGBA{grey, grey, grey, 0xFF}
    }
    black := color.NRGBA{0x00, 0x00, 0x00, 0xFF}
    white := color.NRGBA{0xFF, 0xFF, 0xFF, 0xFF}
    raw := image.NewNRGBA(image.Rect(0, 0, size, size))
    xDirRight := true
    yDirUp := true
    steps := 1
    xToGo := 1
    yToGo := 1
    xHere := (size - 1) / 2
    yHere := size / 2
    for i := 1; i < len(tab); i++ {
        if isGreyscale {
            raw.Set(xHere, yHere, colorTab[tab[i]])
        } else if tab[i] == 1 {
            raw.Set(xHere, yHere, black)
        } else {
            raw.Set(xHere, yHere, white)
        }

        if xDirRight && yDirUp { //Going right
            xHere++
            xToGo--
        } else if yDirUp { //Going Up
            yHere--
            yToGo--
        } else if xDirRight { //Going Down
            yHere++
            yToGo--
        } else { //Going Left
            xHere--
            xToGo--
        }
        //handle direction changes
        if xToGo == 0 {
            xDirRight = !xDirRight
            xToGo = steps + 1
        } else if yToGo == 0 {
            yDirUp = !yDirUp
            steps++
            yToGo = steps
        }
    }
    return raw
}

//Writes an image as a PNG
func writePng(raw image.Image, name string) error {
    file, err := os.Create(name)
    defer file.Close()
    if err != nil {
        return err
    }
    return png.Encode(file, raw)
}

type task struct {
    start  int
    end    int
    primes []int
}

//Counts prime factors of numbers for the range of tab, storing the count in that
//position in tab.  Splits up work into NUM_THREAD goroutines.
func factCount(tab []uint8) int {
    report := make(chan []int) //Go threads report when done, and primes found
    //storage for prime numbers
    primes := make([]int, 0, 20) //default
    if len(tab) > 56 {           //wikipedia claims this is max primes for numbers > 55
        floatlen := float64(len(tab) - 1)
        primes = make([]int, 0, int(floatlen/(math.Log(floatlen)-4)))
    }
    tab[1], tab[2], tab[3] = 1, 1, 1 //initial values
    primes = append(primes, 2)
    primes = append(primes, 3)

    //Start initial jobs
    work := make(chan task)
    for i := 4; i < 8; i++ {
        seed := new(task)
        seed.start = i
        seed.end = i + 1
        seed.primes = primes[:]
        go goFactor(work, tab[:], report)
        work <- *seed
    }
    jobsOut := 4
    queue := make([][]int, 0, 4)   //Holds results recieved out of order
    expected := [4]int{4, 5, 6, 7} //Stores expected result start indexes
    expI := 0                      //The index in expected of the next job
    start := 8                     //where to start next job inclusive
    end := 10                      //where to end next job exclusive
    for jobsOut+len(queue) > 0 {
        if start >= len(tab) { //we're on the last jobs
            if len(queue) > 0 {
                primes = append(primes, queue[0][2:]...)
                queue = queue[1:]
            } else { //there's still jobs out
                results := <-report
                jobsOut--
                primes = append(primes, results[2:]...)
            }
        } else if jobsOut+len(queue) < 4 { //Need to send out a new job
            job := new(task)
            job.start = start
            job.primes = primes[:]
            if end <= len(tab) { //everything good
                job.end = end
            } else { //part of it is hanging off the end
                job.end = len(tab)
            }
            work <- *job
            jobsOut++
            expected[expI] = start //expect these results later
            expI = (expI + 1) % 4  //expect the next job now
        } else {
            //check for queued results that can go through
            clearedQueued := false
            for i, qresult := range queue { //search queue
                if qresult[0] == expected[expI] { //time to clear it
                    clearedQueued = true
                    primes = append(primes, qresult[2:]...)
                    start = qresult[0] * 2
                    end = qresult[1] * 2
                    //overwrite cleared item with last item and shorten queue.
                    queue[i] = queue[len(queue)-1]
                    queue = queue[:len(queue)-1]
                    break
                }
            }
            if !clearedQueued { //Didn't clear anything from queue, wait for a job to finish
                results := <-report
                jobsOut--
                if results[0] == expected[expI] { //results came in order
                    primes = append(primes, results[2:]...)
                    start = results[0] * 2
                    end = results[1] * 2
                } else { //recieved results out of order, place in queue
                    queue = append(queue, results)
                }
            }
        }
    }
    //kill all the threads
    for i := 0; i < 4; i++ {
        work <- *new(task)
    }
    return len(primes) //the number of primes we found
}

//Calculates the factors of number from start (inclusive) to end (exclusive).
//Stores results in tab and reports primes found to the report channel. report[0:2]
//is set to start and end, so that these results can be used in the correct order
func goFactor(work chan task, tab []uint8, report chan []int) {

    for {
        thisJob := <-work
        start := thisJob.start
        end := thisJob.end
        primes := thisJob.primes
        if start == 0 {
            return
        }

        found := make([]int, 2, 2+end-start) //store primes found
        found[0] = start                     //indentifies the results
        found[1] = end
        for i := start; i < end; i++ {
            //max is smallest int > sqrt(i)
            max := int(1 + math.Sqrt(float64(i)))
            for _, f := range primes { //try all primes passed to us
                if f >= max { //it's prime!
                    tab[i] = 1
                    found = append(found, i)
                    break
                }
                if i%f == 0 { //it's not prime 
                    tab[i] = 1 + tab[i/f]
                    break
                }
            }
        }
        report <- found[:] //we're done
    }
}

