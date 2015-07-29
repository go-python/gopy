import hi

print "--- doc(hi)..."
print hi.__doc__

print "--- doc(hi.Hi)..."
print hi.Hi.__doc__

print "--- hi.Hi()..."
hi.Hi()

print "--- doc(hi.Hello)..."
print hi.Hello.__doc__

print "--- hi.Hello('you')..."
hi.Hello("you")

print "--- doc(hi.Add)..."
print hi.Add.__doc__

print "--- hi.Add(1, 41)..."
print hi.Add(1,41)

print "--- hi.Concat('4', '2')..."
print hi.Concat("4","2")

print "--- doc(hi.Person):"
print hi.Person.__doc__

print "--- p = hi.Person()..."
p = hi.Person()
print dir(p)
print "--- p:", p

print "--- p.Name:", p.Name
print "--- p.Age:",p.Age

print "--- doc(hi.Greet):"
print p.Greet.__doc__
print "--- p.Greet()..."
print p.Greet()

print "--- p.String()..."
print p.String()

print "--- doc(p):"
print p.__doc__

print "--- p.Name = \"foo\"..."
p.Name = "foo"

print "--- p.Age = 42..."
p.Age = 42

print "--- p.String()..."
print p.String()
print "--- p.Age:", p.Age
print "--- p.Name:",p.Name

