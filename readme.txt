This application designed to count number of requests for the last N seconds. For each HTTP request it responds
with a number of requests performed since current request time, N seconds back (moving window.
	By default it responds with a number of requests performed 60 seconds back. Every 10ms server persists requests
counter to given file (storage.txt by default). When server shuts down it also saves the requests counter to file.
Data represented with the number requests per second separated by comma. At the start, application will try to restore
counters by reading given file (storage.txt), if it fails the new counter will be created.