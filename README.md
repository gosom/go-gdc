# go-gdc

Add Description here :)

### Usage

Tested with go version go1.16.4 linux/amd64

```
go mod download
go build
```

#### Command line

```
go-gdc registration-number --regNum 81533 --format json
```

#### Rest API

```
go-gdc api --bind :8000
```

```
curl http://localhost:8000/registration-number-search?regNum=81532
```

##### Sample Api Response

```
{
  "Url": "https://olr.gdc-uk.org/SearchRegister/SearchResult?RegistrationNumber=81532",
  "Error": null,
  "StatusCode": 200,
  "Individual": {
    "Name": "David William Suitor",
    "RegistrationNumber": "81532",
    "Status": "Registered",
    "RegistrantType": "Dentist",
    "FirstRegisteredOn": "10 Jan 2003",
    "CurrentPeriodFrom": "10 Jan 2003",
    "CurrentPeriodUntil": "31 Dec 2021",
    "Qualifications": [
      "BDentSc Dubl 2002"
    ]
  }
}
```





