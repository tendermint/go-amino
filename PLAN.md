# Explain why Amino.

* better mapping to OOP languages than proto2/3; aka "proto3/4 Any wants to be
  Amino" --> now is the time to try different approaches.  we already have a
way that work, and the usage of codecs to register types at the app level is
ultimately a necessary interface.  see
https://developers.google.com/protocol-buffers/docs/proto3#any
"https://developers.google.com/protocol-buffers/docs/proto3#any".
* go-amino specifically written such that code serves as spec (better for
  immutable code), and any determinism enforced, etc.
* faster prototype -> production cycle, future compat w/ proto3 fields.
  (not fully supported yet in Amino).

# TODOs

* `genproto/*` to generate complementary proto schema files (for support in other languages)
* use `genproto/*` generated tooling to encode/decode.
  * [ ] use both fuzz tests to check for completeness.
  * [ ] automate the testing of gofuzz tests.


# NOTES

* Golang code generation examples from google:
  https://github.com/golang/protobuf/blob/master/protoc-gen-go/generator/generator.go,
superceded by `google.golang.org/protobuf/compiler/protogen`,
https://github.com/protocolbuffers/protobuf-go/blob/v1.23.0/cmd/protoc-gen-go/internal_gengo/main.go

The primary function is the `<GeneratedFile>.P()` function, which just prints:

 > P prints a line to the generated output. It converts each parameter to a string following the same rules as fmt.Print. It never inserts spaces between parameters.

This is fine for Go code as it can be re-formatted with gofmt.
What Amino's genproto does is in the other direction, and the result is a proto3 file.

The takeaway for code templating is this:  Don't use golang's default templates system, and use something like `.P()` to print, or `.Pl()` to print a line.
`.I(...)` for indentation, such as:

```
g.Pl("prints a line")
	.WithIndent(func(g Generator) {
		g.Pl("prints an intented line")
		g.Pl("prints another intented line")
	})
	.Pl("prints an unintented line")
```

* Some parts of the code appears to use Go's AST to generate or modify code.
  See
https://github.com/protocolbuffers/protobuf-go/blob/d165be301fb1e13390ad453281ded24385fd8ebc/compiler/protogen/protogen.go#L1129

* I'm not aware of anything yet that generates proto files, but here's one struct representation of one:
https://github.com/protocolbuffers/protobuf-go/blob/d165be301fb1e13390ad453281ded24385fd8ebc/compiler/protogen/protogen.go#L463


# Obligatory Morty

```
                                                                                                                                                                                                        
                                                                                                                                                                                                        
                                                                                                  ``                                                                                                    
                                                                                                  //:.                                                                                                  
                                                                                                 `+.-:/.                                                                                                
                                                                                                 /-...-//`              `-                                                                              
                                                                               `.`              `o......:+`           `-/o`                                                                             
                                                                               `+::-.`          /:.......:o`        `-/:.o`                                                                             
                                                                                .+.-:/:-.`     `o.........:+     `-:/-...o`                                                                             
                                                                                 /:...--:/:-.` /:..........+- `-:/:-.....o`                                                                             
                                                                                 `o.......-::/:/...........-o::--........s                                                                              
                                                                                  -+..........-......-::::::::--........-o                                                                              
                                                                                   o-............-://::-----:::///:-..../:    ``..                                                                      
                                                                                   .o.........-:/:-...............::+:-.o...-:/:/:                                                                      
                                                                                    /:.......:/:.....................:+:+:::-...o`                                                                      
                                                                                    `+.....-//........................./+......:+                                                                       
                                                                            ```````..::...-+-..-/:-.....................//....:o`                                                                       
                                                                      `.--:::::::::--.....+-...:o+//-....................s...:+`                                                                        
                                                                      `//-...............+-......:+o+++:.................o--+:`                                                                         
                                                                        `:/-............:/.....:+//:/++/++/-............-+ss.                                                                           
                                                                          `/+-..........+....-o/......-/++////////////////:+:                       -:.`                                                
                                                                            `/+-.......-+....y/////++:-..-h+++o+++++++/+yho/`                       s./+`                                               
                                                                              `//......o-...//     -:.:://s:.o/..........y:-//`                     o...o.                                              
                                                                                `/-....s....:+           `s::s:::::+///+:-h.../:                    :/...s`                                             
                                                                            `-:::.....:+.....+/          :o.-+     :    ./d..::-                    `o...-o                                             
                                                                         `://-........+-.....-/+/:-...:/+/..-s           /o//.                       o....s`                                            
                                                                      .-::-.......-/::+-.......:////::-.....:h+.       `oy:`                         o....+:                                            
                                                                     `::::/:-....//.........................+/./+//::/+y-`                           o....:+                                            
                                                                        ````.-:/.o..........................s-...:///:+-                        `.--.+....-+                                            
                                                                            `/:-.o.......:-............:...-s.........s`                   -////s:--:o:...-+`                                           
                                                                          `-/-...-+:...:/-............./+..s-........-+                   //....-o-...+...-://`                                         
                                                                          :+///::o:-//:o-.--:::::::--...:++:.........o.                   s.......:....-..-..-o                                         
                                                                          `     .+...-o-+........---::/::::-.........+                    o.-.................s`                                        
                                                                               `o-...-y-.-..............---://:::-...+                    :o..................s.                                        
                                                                               /+::::--o.......................-----.s`                   `s..................s`                                        
                                                                         ``````:..``   /+.........................-:/+                    -o.................-o                                         
                                                                      .-::------:::--.-+:+:.....................::--.`                    -+................-o`                                         
                                                                    .::.           `-+o...:+:-...............-:/.`                        .s...............:o.                                          
                                                                  `::`           `-/:o/.....-://+:--------:/:-.                           -y:............-+/`                                           
                                                                `:/.           `-+o--:s-......./s/:----...`                              -s-o/.........-/+-                                             
                                                              `-/-            -::+----:o+:--:/+:+//-/-`                                 -/o--/++::-::://-`                                              
                                                             ./:            ./-.+--------://:---/./- -/-`                             `:: o-----://yo+`                                                 
                                                           `:/`            :/` o-----------------+ /-  -+.                           `/-  `+:-----o:.s`                                                 
                                                         `-+-            `/-  /:-----------------o  +.  `+:`                        `/.     :+:--+-.-o.                                                 
                                                        `+/            `.+`  `o------------------+.  o.   :+`                      .+`        .:+///s/                                                  
                                                      `:+.           ``.+`   +:------------------:/  `o`   .+-                    -+              `/-                                                   
                                                     .+-           `` -+    `s--------------------o   `o     //`                `:/              -/`                                                    
                                                   `//           `.` -+     :/--------------------o    .o     -o.              `/:              /:                                                      
                                                 `-/`           -.  ./      o---------------------s     :/      +/`           `+-             `+.                                                       
                                                ./.           .:`  `/       s---------------------o`     +-      -+.         `+.             :+`                                                        
                                              `/:           `/-    --      -+---------------------o.      o`       /:       `o`             +:                                                          
                                            `:/            /+       +.     +:---------------------+-      `+        .:`    `+`            .o`                                                           
                                           -/`           -/+.       `o     s----------------------o/      -/          -.  `:             //`                                                            
                                         .:.           `/-`+         .+    s---------------------+/+     /:            `.`-            `+-                                                              
                                       .:.            -:` `+        .:-   `s--------------------:+-+   `s-..             .            -/`                                                               
                                      -:            .:`   -:      -/-     .o--------------------s--o  ./:: `-.                       ::`                                                                
                                     .:         `  -.`    /`     /.       -+-------------------+/..o  --`+  `.:.                   ./.                                                                  
                                     /`         `--`      o      /`       //-------------------s...s   .:+    `::`                -:`                                                                   
                                     ./          `-`      o      `+`      +:------------------o/...s    `o.     ./:             `/-`                                                                    
                                      :-           :.     o       .+      o:------------------s....s     `:.     `-/-          ./.                                                                      
                                      `/-           :.   `+        -/     o------------------o:....s      `/       `-/.       :/`                                                                       
                                       `/-           :-  .+         /:    s-----------------:s.....s      +.         `-/-` `./-                                                                         
                                        `::           :- -/          +.   s-----------------o:.....s     /s`           `.-:-.`                                                                          
                                          -/           -:::          `o`  s-----------------s......s`   -/:.                                                                                            
                                           -/`          -s-           `o  y----------------o:......s`  .+ .:                                                                                            
                                            .+`          -/            .+`y----------------s.......o. `o` `+                                                                                            
                                             `+.          ./`           :/s---------------+/.......+- +.   +                                                                                            
                                              `/-          `/`           +o---------------s........+::-    +                                                                                            
                                                :/          `:`          :+--------------/+......../+:     /`                                                                                           
                                                 -/`          :.         //--------------o.........:s      -:                                                                                           
                                                  .+`          -.        o---------------+.........:+      `+                                                                                           
                                                   `+.          /..      s--------------+-.........-o       o                                                                                           
                                                    `/:       .+/s`      y--------------+...........s       o                                                                                           
                                                      :/    `//.//      `s-------------o-...........s       +`                                                                                          
                                                       -+  -+-.//      ..s------------+/............s       /.                                                                                          
                                                        `o:.-:+-       `ss++++oooossyhs.............s       -:                                                                                          
                                                         +.`-:          :+----mmmmmmmmy.............s`      `+                                                                                          
                                                         +-`             s----Nmmmmmmmy.............s`       o                                                                                          
                                                         +.              /dhhhdddddddho.............+-       o                                                                                          
                                                         o`              /dyyyyyyyyyyyy.............//       o`                                                                                         
                                                         o`              yhyyyyyyyyyyyd.............-o       /-                                                                                         
                                                         s              +dyyyyyyyyyyyyd-.............s       -/                                                                                         
                                                         s            `shhyyyyyyyyyyyyhs.............y       `o                                                                                         
                                                         s          `-odyyyyyyyyyyyyyyyh+............s`       s                                                                                         
                                                        `s        ````ohyyyyyyyyyyyyyyyyh:.........../:       o`                                                                                        
                                                        `s            hhyyyyyyyyyyyyyyyyhy...........-o       +.                                                                                        
                                                        .o           -dyyyyyyyyhhhhyyyyyyho...........s       /-                                                                                        
                                                        -/           s:-------/hhhyyyyyyyyd:..........o.      ./                                                                                        
                                                        /:          `m/........+hyyyyyyyyyhh..........:+      `o                                                                                        
                                                        o.          +dy.........shyyyyyyyyyho..........s       s                                                                                        
                                                        s          `dhd-........-yhyyyyyyyyyd:.........o-      o`                                                                                       
                                                        s          /dyh+........./hyyyyyyyyyhh-........-o      +.                                                                                       
                                                       `o         `dhyhy..........ohyyyyyyyyyhs.........s`     ::                                                                                       
                                                       :/         ohyyyd-..........shyyyyyyyyyd/........:+     `+                                                                                       
                                                       +.        -dyyyyho..........-hhyyyyyyyyhh-........s`     o                                                                                       
                                                       s        `yhyyyyyh...........:hyyyyyyyyyhs........:/     o                                                                                       
                                                      `o        +hyyyyyyd:...........+hyyyyyyyyyd+........s`    +`                                                                                      
                                                      -/       -dyyyyyyyhs............shyyyyyyyyyd-.......:+    :.                                                                                      
                                                      +.      .hhyyyyyyyyd.............yhyyyyyyyyhh........o`   .:                                                                                      
                                                      o      `sdyyyyyyyyyd/............-dhyyyyyyyyho.......-+   `+                                                                                      
                                                     `+      o-hyyyyyyyyyhh.............+dyyyyyyyyyh:......./-   +                                                                                      
                                                     /.     o:.ohyyyyyyyyyd-.............shyyyyyyyyyh-.......+   /                                                                                      
                                                     +     +:..:dyyyyyyyyyhs..............yhyyyyyyyyhy.......-/  /`                                                                                     
                                                    `/    +:....hyyyyyyyyyhd..............-hhyyyyyyyyho.......:. .-                                                                                     
                                                    :.   /:.....shyyyyyyyyym:..............-hhyyyyyyyyd:......./ `:                                                                                     
                                                    /   /-....../hyyyyyyyyydy///////////////odyyyyyyyyhh-.......: /                                                                                     
                                                   `: `/-.......-dyyyyyyyyyhs````````````````/hyyyyyyyyhh/////:--.:                                                                                     
                                                   :``/-...-::///hyyyyyyyyyyy                `yhyyyyyyyyd.````.-://`                                                                                    
                                                  `:./:::::-```` yyyyyyyyyyyh`                /hyyyyyyyyh+       `:.                                                                                    
                                                  .//:.``        syyyyyyyyyyd`                .dyyyyyyyyhs        ``                                                                                    
                                                  .`             syyyyyyyyyyd`                `hyyyyyyyyhy                                                                                              
                                                                 syyyyyyyyyyd`                 yhyyyyyyyhh                                                                                              
                                                                 shyyyyyyyyyd`                 ohyyyyyyyyd                                                                                              
                                                                 shyyyyyyyyyd`                 ohyyyyyyyyd                                                                                              
                                                                 shyyyyyyyyyh`                 shyyyyyyyhh                                                                                              
                                                                 yhyyyyyyyyyy                  yyyyyyyyyhs                                                                                              
                                                                 yyyyyyyyyyys                  yyyyyyyyyh+                                                                                              
                                                                 hyyyyyyyyyyh`                 hyyyyyyyyd:                                                                                              
                                                                `dyyyyyyyyyyh                 `dyyyyyyyyd.                                                                                              
                                                                `dyyyyyyyyyhy                 `dyyyyyyyyd`                                                                                              
                                                                .dyyyyyyyyyho                 .dyyyyyyyyh`                                                                                              
                                                                -dyyyyyyyyyh/                 -dyyyyyyyhy                                                                                               
                                                                :dyyyyyyyyyd-                 :dyyyyyyyho                                                                                               
                                                                /hyyyyyyyyyd.                 +hyyyyyyyd/                                                                                               
                                                                +hyyyyyyyyyd`                 ohyyyyyyyd-                                                                                               
                                                                shyyyyyyyyhh                  shyyyyyyyd`                                                                                               
                                                                yhyyyyyyyyhs                  yhyyyyyyyd                                                                                                
                                                                hyyyyyyyyyh+                  dyyyyyyyhy                                                                                                
                                                                dyyyyyyyyyd:                 `dyyyyyyyho                                                                                                
                                                               `dyyyyyyyyyd.                 `myyyyyyyh/                                                                                                
                                                               `dyyyyyyyyyd`                 .dyyyyyyyd.                                                                                                
                                                               -dyyyyyyyyyh                  -dyyyyyyyd`                                                                                                
                                                               :hyyyyyyyyhs                  /hyyyyyyyy                                                                                                 
                                                               +hyyyyyyyyh+                  +hyyyyyyys                                                                                                 
                                                               ohyyyyyyyyh:                  oyyyyyyyh/                                                                                                 
                                                               ohhhhhhhhyy`                  /hhhhyyyh-                                                                                                 
                                                               `/:----..-/                   `o.-/+osy`                                                                                                 
                                                                :-      ./                    o     `/                                                                                                  
                                                               .//---:::++`                   +.--:/ohs:`                                                                                               
                                                              `hdddddddddds.                 `hhddddddddh+.                                                                                             
                                                              oddddddddddddd/`               oddddddddddddds:`                                                                                          
                                                             .dddddddddddddddy-             .mdddddddddddddddy/.                                                                                        
                                                             sddddddddddddddddd+`           omdddddddddddddddddho-`                                                                                     
                                                            .mddddddddddddddddddy-`         yddddddddddddddddddddds-`                                                                                   
                                                            smddddddddddddddddddddo.        dddddddddddddddddddddddds-`                                                                                 
                                                            /++++++++++++++++++++++:       `rippedfromkineticsqurrel/`                                                                                  
                                                                                                                                                                                                        
```
