#!/usr/local/bin/gnuplot

set term pdf size 3.8,2
set output "duration.pdf"
set ylabel "Duration (s)"
set xlabel "Tree dimension"
set notitle
set yrange [0:120]
set xrange [8:12000]
set logscale x

plot "10m" using 1:($2/1000.0) notitle with lines lw 2, \
     "10m" using 1:($2/1000.0):($3/1000.0) notitle with yerrorbars lc 1 lw 2