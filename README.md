# Go LZMA Decoder

This is a Go implementation of the LZMA format, this library contains only the decoder, 
the encoder can be found [here](https://github.com/giovanibageston/golzmaenc).

The library is a port from the Java LZMA SDK.

## Usage

### Importing the library
To import the library use the following import path:
> import "github.com/giovanibageston/golzmad"
 
### Main Library Functions
#### Create Decoder:
> *golzmad*.**NewDecoder**() -> *Decoder*

Returns a new Decoder object.

#### Decode:

> *golzmad*.**Decode**(*input*, *output*, *outSize*, **decodeHeader**) -> bool, *error*

Decodes the data on the input io.Reader and writes the result to the output io.Writer.

If the parameter *decodeHeader* is set to *true* the function will read the LZMA header from the input io.Reader before decoding the data.

Parameter *outSize* is the size of the output buffer, if the parameter *decodeHeader* is set to *true* the parameter outSize will be overwritten with the size contained in the LZMA header.


#### SetDecoderProperties:

Sets the parameters of the decoder.

> *golzmad*.**SetDecoderProperties**(*lc*, *lp*, *pb*, *dictionarySize*) -> *bool*

These must be the ones used to encode the file.

- lc = number of literal context bits, default=3, range=[0,8]
- lp = number of literal pos bits, default=0, range=[0,4]
- pb = number of pos bits, default=2, range=[0,4]
- dictionarySize = dictionary size in bytes, default= 2**23, range=[1, 2**28], must be a power of 2.

### Use Explained
First call the function **NewDecoder** to instantiate the object:

After the decoder was instantiated you have two options: 

-You can call the function **Decode** and let it read the LZMA header from the io.Reader

-You can call the function **SetDecoderProperties** to set the decoder properties and then call **Decode**.

### Calling Decode WITHOUT setting the decoder properties
In this case, parameter *decodeHeader* must be *true*, this will make the Decoder read the LZMA header from the input, in this case the parameter outSize will be overwritten with the size contained in the LZMA header. 

### Calling Decode AFTER setting the decoder properties
In this case call **SetDecoderProperties** to set the decoder properties and then call **Decode** with the parameter *decodeHeader* set to *false*.

in this case the parameter outSize will be used as the size of the output buffer.
