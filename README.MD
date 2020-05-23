# [Kongjie Spider](src/KongjieSpider/main/kongjie.go)
A concurrent spider using goroutines to crawl images from kongjie.com.

This spider uses golang internal library `net/http` to get html content, `regexp` to filter elements and `goroutine` to crawl concurrently. Blog is here: [Go语言进阶之路：并发爬虫，爬取空姐网所有相册图片](https://blog.csdn.net/c315838651/article/details/105895186).

Running state:

![kongjie spider](images/kongjie_spider.png)

# [LRU(least recently used) Cache](src/LRUCache/LRU.go)
A thread-safe LRU cache tool implemented by golang. 

This tool is inspired by LinkedHashMap of Java. When we set capacity of our LRU cache, least recently unused elements will removed from cache if it reaches its capacity.

Blog is here: [Go语言进阶之路：手撸一个LRU缓存](https://blog.csdn.net/c315838651/article/details/105741886).

# BlockChain in Golang

A simple blockchain application implemented in Golang. Anyone can get started with Golang by doing this project, as well as by following the tutorial in reference below.

As the tutorial shows, what you will be able to do with this project:

1. Create your own blockchain
2. Understand how hashing works in maintaining integrity of the blockchain
3. See how new blocks get added
4. See how tiebreakers get resolved when multiple nodes generate blocks
6. View your blockchain in a web browser
7. Write new blocks
8. Get a foundational understanding of the blockchain so you can decide where your journey takes you from here!
9. Run a terminal that provides a genesis block
10. Fire up as many additional terminals as you like and have them write blocks to the first terminal
11. Have the first terminal broadcast updated blocks to the new terminals

# Reference
1. [Code your own blockchain in less than 200 lines of Go](https://medium.com/@mycoralhealth/code-your-own-blockchain-in-less-than-200-lines-of-go-e296282bcffc)
