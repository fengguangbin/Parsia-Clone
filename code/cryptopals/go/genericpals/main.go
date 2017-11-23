// Collection of utilities

package genericpals

import (
    "encoding/hex"
    "encoding/base64"
    "errors"
    "fmt"
    "bytes"
    "sort"
    "os"
    "bufio"
    "strings"
    "math/bits"
    "crypto/aes"
    // "reflect"
)

// Unhexlify converts an ASCII-Hex encoded string to the corresponding []byte
// Params:
//      hexString: string containing bytes
//
// Return: []byte converted from ASCII-Hex
func Unhexlify(hexString string) []byte {
    hexlified, err := hex.DecodeString(hexString)
    if err != nil {
        panic(err)
        return nil
    } else {
        return hexlified
    }
}

// Hexlify converts a []byte to ASCII-Hex
// Params:
//      inputBytes: []byte containing ASCII-Hex bytes
//
// Return: string containing ASCII-Hex representation of []byte
func Hexlify(inputBytes []byte) string {
    return hex.EncodeToString(inputBytes)
}

// B64EncodeByteToStr encode a []byte in base64 standard
// Params:
//      inputBytes: []byte that will be encoded to base64
//
// Return: string containing base64 encoded data
func B64EncodeByteToStr(inputBytes []byte) string {
    return base64.StdEncoding.EncodeToString(inputBytes)
}

// B64DecodeStrToByte decodes a string and returns the result in []byte
// Params:
//      inputString: base64 encoded string
//
// Return: []byte containing decoded bytes
func B64DecodeStrToByte(inputString string) ([]byte, error) {
    decoded, err := base64.StdEncoding.DecodeString(inputString)
    if err != nil {
        return nil, err
    }
    
    return decoded, nil
}

// SameLengthXOR XORs two []byte of same length
// Params:
//      inputBytes1, inputBytes2: []byte containing strings to be XOR-ed
//
// Return: XOR result and error if applicable
func SameLengthXOR(inputBytes1 []byte, inputBytes2 []byte) ([]byte, error) {
    if len(inputBytes1) != len(inputBytes2) {
        errorString := fmt.Sprintf("Input1 length %d != %d Input2 length",
                                    len(inputBytes1), len(inputBytes2))
        return nil, errors.New(errorString)
    } else {
        // Make a slice because Go does not have dynamic-length arrays
        outputBytes := make([]byte, len(inputBytes1))
        for i:=0; i<len(inputBytes1); i++ {
            outputBytes[i] = inputBytes1[i] ^ inputBytes2[i]
        }

        return outputBytes, nil
    }
}

// ByteArrayEqual return true if []bytes are of the same length and have the
// same bytes
// Params:
//      inputBytes1, inputBytes2: []bytes to be compared
//
// Return: true if equal, otherwise false
func ByteArrayEqual(inputBytes1 []byte, inputBytes2 []byte) bool {
    return bytes.Equal(inputBytes1, inputBytes2)
}

// SingleByteXOR XORs a single byte key with a []byte
// Params:
//      key: Single byte key
//      inputBytes: []byte to be XOR-ed
//
// Return: XOR-ed []byte
func SingleByteXOR(inputBytes []byte, key byte) []byte {
    outputBytes := make([]byte, len(inputBytes))

    for i:=0; i<len(inputBytes); i++ {
        outputBytes[i] = inputBytes[i] ^ key
    }

    return outputBytes
}

// isPrintableASCII returns true if input's ASCII-Hex code is
// between 0x20 [inclusive] and 0x7F. See an ASCII table for more info.
// Trust me, I am a pseudo reverse engineer
// Params:
//      char: byte char to be evaluated
//
// Return: true if char is printable, false if not
func isPrintableASCII(char byte) bool {
    return ((0x20 <= char) && (char < 0x7F))
}

// English letter frequency array
// Cannot have const maps in go
var frequency = map[string]float64 {
    "a": 0.0651738,
    "b": 0.0124248,
    "c": 0.0217339,
    "d": 0.0349835,
    "e": 0.1041442,
    "f": 0.0197881,
    "g": 0.0158610,
    "h": 0.0492888,
    "i": 0.0558094,
    "j": 0.0009033,
    "k": 0.0050529,
    "l": 0.0331490,
    "m": 0.0202124,
    "n": 0.0564513,
    "o": 0.0596302,
    "p": 0.0137645,
    "q": 0.0008606,
    "r": 0.0497563,
    "s": 0.0515760,
    "t": 0.0729357,
    "u": 0.0225134,
    "v": 0.0082903,
    "w": 0.0171272,
    "x": 0.0013692,
    "y": 0.0145984,
    "z": 0.0007836,
    " ": 0.1918182,
}

// printableEnglish returns a score based on closeness of the input string to
// the English language
// All chars are parsed, for each char with a value in the frequency map, the
// score is increased by the frequency value
// Params:
//      textBytes: Input string in []byte
//
// Returns: Score in float32
func PrintableEnglish(textBytes []byte) float64 {

    var score float64

    for _, char := range textBytes {
        ch := strings.ToLower(string(char))

        if frequency[ch] != 0 {
            score += frequency[ch]
        } else {
            score -= 0.01
        }
    }

    return score
}

// Struct to hold the results
type Result struct {
    Key byte
    PrintableScore float64
    Plaintext string
}

// BreakSingleByteXOR attempts to break single byte XOR encryption.
// It uses printableEnglish to find the one with the highest score and
// returns the struct
// Params:
//      ciphertext: []byte containing single-byte XOR encrypted string
//
// Return: genericpals.Result
func BreakSingleByteXOR(ciphertext []byte) Result {

    results := make([]Result, 0x100)

    for key:=0; key<0x100; key++ {

        plaintext := SingleByteXOR(ciphertext, byte(key))

        results[key].Plaintext = string(plaintext)
        results[key].Key = byte(key)
        results[key].PrintableScore = PrintableEnglish(plaintext)

    }

    // Sorting the array using sort.Slice
    // https://stackoverflow.com/a/42872183

    // Sort
    sort.Slice(results[:], func(i, j int) bool {
        return results[i].PrintableScore > results[j].PrintableScore
    })

    return results[0]
}


// ReadLines reads lines from a text file and stores each in an element of a
// string array
// Params:
//      filePath: string containing path to file
//
// Return: []string containing each line in one element and error if applicable.
// How to read lines using go
// https://stackoverflow.com/a/16615559
// "Scanner does not deal well with lines longer than 65536 characters."
func ReadLines(filePath string) ([]string, error) {
  
  inputFile, err := os.Open(filePath)
  if err != nil {
    return nil, err
  }

  // Close the file after function returns
  defer inputFile.Close()

  var textLines []string

  scanner := bufio.NewScanner(inputFile)

  for scanner.Scan() {
    textLines = append(textLines, scanner.Text())
  }

  return textLines, scanner.Err()
}

// XOR XORs a key with a []byte and repeats the key
// Params:
//      key: []byte key
//      inputBytes: []byte to be XOR-ed
//
// Return: XOR-ed []byte
func XOR(inputBytes []byte, key []byte) []byte {
    outputBytes := make([]byte, len(inputBytes))

    keyLen := len(key)

    for i:=0; i<len(inputBytes); i++ {
        outputBytes[i] = inputBytes[i] ^ key[i % keyLen]
    }

    return outputBytes
}

// HammingDistance calculates the hammind distance between two []byte strings
// XOR two bytes and then count the 1 bits in result
// Params:
//      stringBytes1, stringBytes2: []byte
//
// Return: Hamming distance as int
func HammingDistance(stringBytes1 []byte, stringBytes2 []byte) int {
    if len(stringBytes1) != len(stringBytes2) {
        panic("Input bytes are not of same length")
        return -1
    }

    dist := 0

    for i := 0; i < len(stringBytes1); i++ {

        // Thanks Travis
        d1 := bits.OnesCount8(stringBytes1[i] ^ stringBytes2[i])

        dist += d1
    }

    return dist
}

// ReadAllFile is similar to ReadLines but returns one string instead of each
// line being in one string
// Params:
//      filePath: string containing path to file
//
// Return: string containing all characters in the file with new lines removed
func ReadAllFile(filePath string) (string, error) {
    inputFile, err := os.Open(filePath)
    if err != nil {
        return "", err
    }

    // Close the file after function returns
    defer inputFile.Close()

    var allLines string

    scanner := bufio.NewScanner(inputFile)

    for scanner.Scan() {
        allLines += scanner.Text()
    }

    return allLines, scanner.Err()
}

// GetTwoSeqBytes returns n-th two sequential blocks of size bytes from input
// Params:
//      inputBytes: input []byte
//      size      : size of selected bytes
//      n         : n-th tuple (how do I say it in English lol?)
//
// Returns: Two sequential []byte of same size
func GetTwoSeqBytes(input []byte, size int, n int) (block1, block2 []byte) {

    // 0 is unnecessary but this will readable 6 months down the road
    block1 = input[(2*n+0)*size:(2*n+1)*size]
    block2 = input[(2*n+1)*size:(2*n+2)*size]

    return block1, block2
}

// ECBAESEncrypt encrypts []byte plaintext to []byte ciphertext
// Params:
//      plaintext: []byte plaintext
//      key      : []byte key
//
// Return: []byte ciphertext and error if any
func ECBAESEncrypt(plaintext []byte, key []byte) ([]byte, error) {

    // Get an AES block
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    ciphertext := make([]byte, len(plaintext))

    // This should be 16 for AES forever and ever
    blockSize := block.BlockSize()

    for i:= 0; i<len(plaintext); i+=blockSize {
        block.Encrypt(ciphertext[i:i+blockSize], plaintext[i:i+blockSize])
    }

    return ciphertext, nil
}

// ECBAESDecrypt decrypt []byte ciphertext to []byte plaintext
// Params:
//      ciphertext: []byte ciphertext
//      key       : []byte key
//
// Return: []byte plaintext and error if any
func ECBAESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {

    // Get an AES block
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    plaintext := make([]byte, len(ciphertext))

    // This should be 16 for AES forever and ever
    blockSize := block.BlockSize()

    for i:=0; i<len(ciphertext); i+=blockSize {
        block.Decrypt(plaintext[i:i+blockSize], ciphertext[i:i+blockSize])
    }

    return plaintext, nil
}

// SplitBytes splits a []byte into equal lengths of n
// Last section may be smaller than the rest.
// Params:
//      inputBytes: []byte
//      n: int - split length
//
// Return: [][]byte
func SplitBytes(inputBytes []byte, n int) [][]byte {
    
    size := len(inputBytes)/n
    if (len(inputBytes) % n) != 0 {
        size += 1
    }

    splits := make([][]byte, size)

    for i:=0; i<len(inputBytes); i+=n {

        if i+n > len(inputBytes) {
            splits[i / n] = inputBytes[i:len(inputBytes)]
        } else {
            splits[i / n] = inputBytes[i:i+n]
        }
    }

    return splits
}

// ByteRepeat creates a []byte by repeating a byte
// It works the same way as bytes.Repeat([]byte, count) but accepts a single byte
// Params:
//      repeatByte: byte to be repeated
//      n: int - number of times to repeat repeatByte
//
// Returns: []byte containing repeatByte n times
func ByteRepeat(repeatByte byte, n int) []byte {

    if n < 0 {
        panic("Negative repeat count")
    }

    output := make([]byte, n)

    for i := 0; i < n; i++ {
        output[i] = repeatByte
    }

    return output
}

// PKCS7Pad pads a []byte to a multiple of blockSize
// Padding value is the number of padded bytes
// For example if we are padding 4 bytes, padding value will be 0x04
// Params:
//      bytesToPad: []byte - unpadded input
//      blockSize : int - input will be padded to a multiple of this number
//
// Return: []byte - padded input
func PKCS7Pad(bytesToPad []byte, blockSize int) []byte {
    
    if len(bytesToPad) == 0 {
        panic("Cannot pad an empty []byte")
    }

    paddingSize := blockSize - (len(bytesToPad) % blockSize)

    // bytes.Repeat needs []byte - we have int
    // Thus we use our own function
    padding := ByteRepeat(byte(paddingSize), paddingSize)

    // Second param of append needs the primitive time of the first param
    // For example in this case bytesToPad is []byte so padding should be byte
    // But because it's []byte, we pass it as padding... to pass the bytes
    // one by one
    // At this point I am not exactly sure how this works other than it works!
    outputBytes := append(bytesToPad, padding...)

    return outputBytes
}

// PKCS7Unpad removes PKCS7 padding from []byte if any
// Reads the last byte, then reads that many bytes. If they are all the same 
// value then we know padding is correct and we will remove it, otherwise error.
// Params:
//      paddedBytes: []byte padded input
//
// Return: []byte unpadded input and error if any
func PKCS7Unpad(paddedBytes []byte) ([]byte, error) {

    paddedLength := len(paddedBytes)

    // Read the last byte
    padding := paddedBytes[paddedLength-1]
    paddingLength := int(padding)

    // Check if we even have enough bytes
    if paddedLength < paddingLength {
        return nil, errors.New("Input is too small to be padded!")
    }

    // Read last n bytes
    for i:=0; i<paddingLength; i++ {
        if paddedBytes[paddedLength-1-i] != padding {

            errorString := fmt.Sprintf("Wrong padding at byte %d." +
                                       "\nExpected %x but got %x.",
                                       padding, paddedBytes[paddedLength-1-i])

            return nil, errors.New(errorString)
        }
    }

    return paddedBytes[:paddedLength-paddingLength], nil
}

// CBCAESDecrypt decrypts []byte from using AES-CBC
// Params:
//      ciphertext: []byte - encrypted data
//      key: []byte - should be aes.BlockSize (16) bytes
//      iv : []byte - Initialization Vector, should be aes.BlockSize (16) bytes
//
// Returns: []byte - decrypted string and error if any
func CBCAESDecrypt(ciphertext, key, iv []byte) ([]byte, error) {

    // 1. Check if ciphertext is a multiple of aes.BlockSize
    if len(ciphertext) % aes.BlockSize != 0 {
        errorString := fmt.Sprintf("Ciphertext is %d bytes which is not" +
                                   "a multiple of %d (aes.BlockSize).",
                                   len(ciphertext), aes.BlockSize)

        return nil, errors.New(errorString)
    }

    // 2. Check if IV and aes.BlockSize are of same length
    if len(iv) != aes.BlockSize {
        errorString := fmt.Sprintf("IV has wrong length." + 
                                   "\nExpected %d, got %d",
                                   aes.BlockSize, len(iv))

        return nil, errors.New(errorString)
    }

    // 3. Split the ciphertext into aes.BlockSize (16) byte blocks
    blocks := SplitBytes(ciphertext, aes.BlockSize)

    // 4.  For each block
    // 4.1 ECBAESDecrypt the block
    // 4.2 XOR with IV
    // 4.3 store plaintext
    // 4.4 IV := block (old ciphertext)

    var plaintext []byte

    for _, block := range blocks {
        // 4.1 ECBAESDecrypt the block
        decryptedBlock, err := ECBAESDecrypt(block, key)

        // Check if it was decrypted correctly
        if err != nil {
            return nil, err
        }

        // 4.2 XOR with IV
        decryptedBlock = XOR(decryptedBlock, iv)

        // 4.3 store plaintext
        plaintext = append(plaintext, decryptedBlock...)

        // 4.4 IV := block (old ciphertext)
        iv = block
    }

    return plaintext, nil
}

