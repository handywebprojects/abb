////////////////////////////////////////////////////////////////

package abb

////////////////////////////////////////////////////////////////

import(	
	"fmt"
	"os"	
	"strconv"
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
