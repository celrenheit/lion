EXAMPLES_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $EXAMPLES_DIR

for example_dir in *
do
    if [ -d "${example_dir}" ]; then
        go build -o $example_dir/built ./$example_dir || { echo "${example_dir} failed" ; exit 1; }
        rm $example_dir/built
    fi
done

echo "PASS EXAMPLES"