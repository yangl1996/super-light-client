#!/usr/local/bin/gnuplot

set term pdf size 3.236,2 font "Serif, 16"
set output "throughput.pdf"
set ylabel "Throughput (games/s)"
set xlabel "# games in parallel"
set notitle
set yrange [0:1000]
set xrange [1:300]

plot "throughput" using 1:($1/$2*1000.0) notitle with lines lw 2
