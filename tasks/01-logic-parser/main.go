package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)


type Token struct{
    Type string // "IDENT", "OP", "PAREN", "TRUE", "FALSE"
    Value string
}

type Parser struct {
    Tokens []Token
    Pos    int // индекс текущего токена
    Vars   map[string]bool
}


func IsEquation(str string) bool {
    re := regexp.MustCompile(`^\s*[a-z]+\s*=\s*(True|False)\s*;\s*$`)
    return re.MatchString(str)
}

func ParseEquation(line string) (string, bool, error) {
    re := regexp.MustCompile(`^\s*([a-z]+)\s*=\s*(True|False)\s*;\s*$`)
    m := re.FindStringSubmatch(line)
    if m == nil {
        return "", false, fmt.Errorf("invalid equation")
    }
    return m[1], m[2] == "True", nil
}

func Tokenize(str string) ([]Token, error) {
    var tokens []Token
    var buf strings.Builder

    for _, symbol := range str {
        switch symbol {
        case '(', ')', ' ':
            if err := FlushToken(&buf, &tokens); err != nil {
                return nil, err
            }
            if symbol == '(' {
                tokens = append(tokens, Token{Type: "LPAREN", Value: "("})
            } 
            if symbol == ')' {
                tokens = append(tokens, Token{Type: "RPAREN", Value: ")"})
            }
        default:
            buf.WriteRune(symbol)
        }
    }
    
    if err := FlushToken(&buf, &tokens); err != nil {
        return nil, err
    }
    return tokens, nil
}

func FlushToken(b *strings.Builder, tokens *[]Token) error {
    token := b.String()
    if token == "" {
        return nil
    }
    switch token {
    case "and", "or", "xor", "not":
        *tokens = append(*tokens, Token{Type: "OP", Value: token})
    case "True":
        *tokens = append(*tokens, Token{Type: "TRUE", Value: token})
    case "False":
        *tokens = append(*tokens, Token{Type: "FALSE", Value: token})
    default:
        for _, r := range token {
            if r < 'a' || r > 'z' {
                return fmt.Errorf("invalid token: %s", token)
            }
        }
        *tokens = append(*tokens, Token{Type: "IDENT", Value: token})
    }
    b.Reset()
    return nil
}





func CalculateExpression(str string, vars map[string]bool) (bool, error) {
    tokens, err := Tokenize(str)
    if err != nil {
        return false, err
    }
    
    p := &Parser{
        Tokens: tokens,
        Pos:    0,
        Vars:   vars,
    }
    
    result, err := p.ParseOr()  
    if err != nil {
        return false, err
    }
    
    if p.Pos != len(p.Tokens) {
        return false, fmt.Errorf("unexpected token: %s", p.Tokens[p.Pos].Value)
    }
    return result, nil
    
}

func (p *Parser) ParsePrimary() (bool, error) {
    tok, ok := p.Current()
    if !ok {
        return false, fmt.Errorf("unexpected end of expression")
    }

    switch tok.Type {
    case "TRUE":
        p.Advance()
        return true, nil
    case "FALSE":
        p.Advance()
        return false, nil
    case "IDENT":
        val, exists := p.Vars[tok.Value]
        if !exists {
            return false, fmt.Errorf("undefined variable: %s", tok.Value)
        }
        p.Advance()
        return val, nil
    case "LPAREN":
        p.Advance()
        result, err := p.ParseOr()   
        if err != nil {
            return false, err
        }
        closing, ok := p.Current()
        if !ok || closing.Type != "RPAREN" {
            return false, fmt.Errorf("expected )")
        }
        p.Advance()
        return result, nil
    default:
        return false, fmt.Errorf("unexpected token: %s", tok.Value)
    }
}




func (p *Parser) ParseNot() (bool, error) {
    tok, ok := p.Current()
    if ok && tok.Type == "OP" && tok.Value == "not" {
        p.Advance()
        result, err := p.ParseNot()
        if err != nil {
            return false, err
        }
        return !result, nil
    }
    return p.ParsePrimary()
}




func (p *Parser) ParseAnd() (bool, error) {
    left, err := p.ParseNot()
    if err != nil {
        return false, err
    }
    for {
        tok, ok := p.Current()
        if !ok || tok.Type != "OP" || tok.Value != "and" {
            break
        }
        p.Advance()
        right, err := p.ParseNot()
        if err != nil {
            return false, err
        }
        left = left && right
    }
    return left, nil
}

func (p *Parser) ParseXor() (bool, error) {
    left, err := p.ParseAnd()
    if err != nil {
        return false, err
    }
    for {
        tok, ok := p.Current()
        if !ok || tok.Type != "OP" || tok.Value != "xor" {
            break
        }
        p.Advance()
        right, err := p.ParseAnd()
        if err != nil {
            return false, err
        }
        left = left != right
    }
    return left, nil
}

func (p *Parser) ParseOr() (bool, error) {
    left, err := p.ParseXor()
    if err != nil {
        return false, err
    }
    for {
        tok, ok := p.Current()
        if !ok || tok.Type != "OP" || tok.Value != "or" {
            break
        }
        p.Advance()
        right, err := p.ParseXor()
        if err != nil {
            return false, err
        }
        left = left || right
    }
    return left, nil
}


func (p *Parser) Current() (Token, bool) {
    if p.Pos >= len(p.Tokens) {
        return Token{}, false
    }
    return p.Tokens[p.Pos], true
}

func (p *Parser) Advance() {
    p.Pos++
}




func main() {
	scanner := bufio.NewScanner(os.Stdin)
    logicOperations := map[string]int{"not": 3, "and": 2, "xor": 1, "or": 0}
    variablesMap := make(map[string]bool)
    
	haveExpression := false

    for scanner.Scan() {
        line := scanner.Text()
    
        if strings.TrimSpace(line) == "" {
            continue   // пустые строки просто пропускаем
        }
    
        if IsEquation(line) {
            name, value, err := ParseEquation(line)
            if err != nil {
                fmt.Println("[error]")
                return
            }
            if _, isOp := logicOperations[name]; isOp {
                fmt.Println("[error]")
                return
            }
            variablesMap[name] = value
            continue
        }
    
        // это выражение
        haveExpression = true
        result, err := CalculateExpression(line, variablesMap)
        if err != nil {
            fmt.Println("[error]")
            return
        }
        if result {
            fmt.Println("True")
        } else {
            fmt.Println("False")
        }
    }

    if !haveExpression {
        fmt.Println("False")
    }
}