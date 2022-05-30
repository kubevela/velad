CHART_FILE=pkg/resources/static/vela/charts/vela-core/Chart.yaml

VERSION_TO=$1

# Works on Mac: see https://stackoverflow.com/questions/2320564/sed-i-command-for-in-place-editing-to-work-with-both-gnu-sed-and-bsd-osx
sed -i "" -e "s/version: v.*/version: $VERSION_TO/g" $CHART_FILE
sed -i "" -e "s/appVersion: v.*/appVersion: $VERSION_TO/g" $CHART_FILE