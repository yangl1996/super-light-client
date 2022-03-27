#for m in 1000 25000 62500 156250 390625 976562 2441406 6103515; do
#for m in 10000000 25000000 50000000; do
for m in 75000000 100000000; do
	pkill testbed
	./testbed exp -generate $m -dim 300
	./testbed exp -serve -dim 300 &
	lastpid=$!
	sleep 15
	../super-light-client verify -dim 300 $(./testbed verify) &> $m.out
	res=`cat $m.out | grep finished | cut -f 7,10 -d' '`
	echo "$m $res" >> output
	kill $!
	wait $!
done
