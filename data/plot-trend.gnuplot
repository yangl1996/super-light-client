#!/usr/local/bin/gnuplot

set term pdf size 3.236,2 font "Serif, 16"
set output "trend.pdf"
set ylabel "Duration (s)"
set xlabel "Ledger size"
set notitle
set yrange [0:120]
set xrange [500:20000000]
set logscale x

plot "trend" using 1:($2/1000.0):($3/1000.0) notitle with errorlines lw 2
