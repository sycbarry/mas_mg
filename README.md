# automate the mas must-gather log process for your instances.

## pre-reqs.

1. docker installed.
2. a readily accessible openshift cluster.

## installation 

1. clone this repo.
2. run ``` go build ```
3. copy over the mas_mg.[exe] file to where you want to run the program.

#### alternatively

1. just download the release package for this repo.
2. extract the mas_mg.[exe] file.

## instructions.

``` ./mas_mg "oc login --token..." ```

> the output file will be called mas-must-gather.zip. upload this to your IBM ticket...


