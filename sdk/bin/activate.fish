#!/usr/bin/env fish

set current_dir (realpath (dirname (status -f)))

exec bash -c ". $current_dir/activate; exec fish"