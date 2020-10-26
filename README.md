# go-learn-opengl

Code translated into go from the [great tutorial](https://learnopengl.com/) by [Joey de Vries](https://twitter.com/JoeyDeVriez). The original repo is [here](https://github.com/JoeyDeVries/LearnOpenGL).

## Building

### Linux specific (well ubuntu but it is mostly the same on other distros likely)

For model loading 
```
sudo apt-get install assimp-utils
```

For sound
```
sudo apt-get install libasound2-dev
```

### Windows specific

For model [go to this page](http://www.assimp.org/index.php/downloads) and download [assimp.3.1.1](https://sourceforge.net/projects/assimp/files/assimp-3.1/) ```assimp-3.1.1-win-binaries.zip```.

Unzip this file and find ```bin64/assimp.dll``` and move this into the ```C:/Windows/System32``` folder. In the repo I included ```dlls/assimp.dll``` for ease.

### Mac specific

For model loading
```
brew install assimp
```

### Building


Download deps
```
go mod download
```

Then it is as easy as going to the folder of the example you would to to run then
```
go run hello_triangle.go
```

Open up an issue if you are having trouble with getting the code to build. 

### Great examples that helped along the way

https://github.com/cstegel/opengl-samples-golang

https://github.com/raedatoui/learn-opengl-golang

https://github.com/tbogdala/assimp-go/blob/master/assimp.go