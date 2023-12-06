Go build -o build/chord
if [ $# -eq 1 ] 
then
  build/chord -a $1
else
  build/chord -j $1 -a $2
fi
