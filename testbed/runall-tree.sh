for m in 7000 8000 9000; do
	pkill testbed
	./testbed exp -generate 10000000 -dim $m
	./testbed exp -serve -dim $m &
	lastpid=$!
	sleep 15
	../super-light-client verify -dim $m $(./testbed verify) &> $m.out
	res=`cat $m.out | grep finished | cut -f 7,10 -d' '`
	echo "$m $res" >> output
	kill $!
	wait $!
done
