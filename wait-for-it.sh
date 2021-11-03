for var in "$@"
do
    until curl -s -f -o /dev/null $var
    do
    sleep 1
    done
done

