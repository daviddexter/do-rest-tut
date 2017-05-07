#!/bin/sh
for i in *
do 
    case $i in
    main.go)        
        go run $i
        break
        ;;    
    esac        
done

