////////////////////////////////////////////////////////////////

package abb

////////////////////////////////////////////////////////////////

import(	
	"fmt"
	"os"	
	"strconv"
	"strings"
)

////////////////////////////////////////////////////////////////

func init(){
	if false{
		fmt.Println("--> initializing utils")
		fmt.Println("--> utils initialized")
	}	
}

////////////////////////////////////////////////////////////////

func Envint(key string, defaultvalue int) int{
	valuestr, haskey := os.LookupEnv(key)
	if haskey{
		intvalue, err := strconv.Atoi(valuestr)
		if err != nil{
			return defaultvalue
		}else{
			return intvalue
		}
	}
	return defaultvalue
}

func Envintarray(key string, defaultvalue []int) []int{
	valuestr, haskey := os.LookupEnv(key)
	if haskey{
		intarray := Str2intarray(valuestr)
		return intarray
	}
	return defaultvalue
}

func Envstr(key string, defaultvalue string) string{
	valuestr, haskey := os.LookupEnv(key)
	if haskey{
		return valuestr
	}
	return defaultvalue
}

////////////////////////////////////////////////////////////////

func Fen2bookletindex(fen string, mod int) int{
	sum := 0
	for i, c := range fen{
		sum += (i+1) * int(c)
	}
	return sum % mod
}

////////////////////////////////////////////////////////////////

func Intarray2str(intarray []int) string{
	strs := []string{}
	for _, intvalue := range(intarray){
		strs = append(strs, strconv.Itoa(intvalue))
	}
	return strings.Join(strs, ",")
}

func Str2intarray(str string) []int{
	strs := strings.Split(str, ",")
	intarray := []int{}
	for _, str := range(strs){
		value, _ := strconv.Atoi(str)
		intarray = append(intarray, value)
	}
	return intarray
}

////////////////////////////////////////////////////////////////