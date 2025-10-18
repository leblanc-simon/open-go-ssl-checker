#!/usr/bin/env bash

CURRENT_DIR=$(dirname "$0")

echo "package translations" > ${CURRENT_DIR}/../translations/fake-trans.go
echo "" >> ${CURRENT_DIR}/../translations/fake-trans.go
echo "//go:generate gotext -srclang=en update -out=catalog.go -lang=en,fr leblanc.io/open-go-ssl-checker/translations" >> ${CURRENT_DIR}/../translations/fake-trans.go
echo "" >> ${CURRENT_DIR}/../translations/fake-trans.go
echo "import \"leblanc.io/open-go-ssl-checker/internal/localizer\"" >> ${CURRENT_DIR}/../translations/fake-trans.go

echo "func FakeTrans () {" >> ${CURRENT_DIR}/../translations/fake-trans.go
echo "	l := localizer.Get(\"en\")" >> ${CURRENT_DIR}/../translations/fake-trans.go

cat ${CURRENT_DIR}/../templates/*.html \
    | grep -Ee '\{\{ *Translate "([^"]+)" *\}\}' \
    | sed 's/{{/\n/g' \
    | sed "s/\(.*\)\(Translate \"[^\"]\+\"\)\(.*\)/\2/" \
    | grep -E '^Translate "([^"]+)"' \
    | sed 's/Translate /    l.Translate(/' \
    | sed 's/"$/"\)/' \
    >> ${CURRENT_DIR}/../translations/fake-trans.go

echo "}" >> ${CURRENT_DIR}/../translations/fake-trans.go

go generate ${CURRENT_DIR}/../translations/fake-trans.go

rm ${CURRENT_DIR}/../translations/fake-trans.go

sed -i 's/package translations/package catalog/' ${CURRENT_DIR}/../translations/catalog.go
mv ${CURRENT_DIR}/../translations/catalog.go ${CURRENT_DIR}/../internal/catalog/catalog.go