#!/usr/local/bin/gnuplot

set term pdf size 3.236,2 font "Serif, 16"
set output "duration.pdf"
set ylabel "Duration (s)"
set xlabel "Tree degree"
set notitle
set yrange [0:120]
set xrange [8:12000]
set logscale x

plot "10m" using 1:($2/1000.0):($3/1000.0) notitle with errorlines lw 2
