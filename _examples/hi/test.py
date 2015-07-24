import hi

print "--- hi.Hi()..."
hi.Hi()

print "--- hi.Hello('you')..."
hi.Hello("you")

print "--- hi.Add(1, 41)..."
print hi.Add(1,41)

print "--- hi.Concat('4', '2')..."
print hi.Concat("4","2")

print "--- doc(hi.Person):"
print hi.Person.__doc__

print "--- p = hi.Person()..."
p = hi.Person()
print p
print dir(p)

print "--- p.Name:", p.Name
print "--- p.Age:",p.Age

print "--- doc(p):"
print p.__doc__

