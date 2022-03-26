for m in 1000 25000 62500 156250 390625 976562 2441406 6103515; do
	pkill testbed
	./testbed exp -generate $m -dim 300
	./testbed exp -serve -dim 300 &
	lastpid=$!
	../super-light-client verify -dim 300 $(./testbed verify) &> $m.out
	kill $!
	wait $!
done
