# Dungeon Maestro

The dungeon maestro application is a slack bot that can respond to D&D related commands. Making slack that much more
useful for perspective campaigns. Only works with 5e

## Actions

### Roll
The roll action allows you to ask the dungeon maestro to roll some arbitrary number of dice of your own dimension! You
can ask the dungeon maestro to roll some dice for you.

#### Request Format
```
/roll <number of dice> d<number of sides on the dice>
```

#### Examples
Whether you are rolling a d20 to see if you successfully charmed the evil wizard to allow your part to escape:
```
/roll 1 d20
roryj rolled 1 d20 and got 17
```

Or rolling for some serious magic damage:
```
/roll 10 d8

roryj rolled 10 d8 and got 44
```

Or you're a cheater and think no one notices that you actually are rolling a d24:
```
/roll 1 d24

roryj rolled 1 d24 and got 21
```

### Spell
The spell action gets you spell information for any spell that can be found on [D&D Beyond](https://www.dndbeyond.com).

### Request Format
```
/spell <spell name>
```

#### Examples
```
/spell hideous laughter


Description of the spell hideous laughter
+----------+-----------+-----------+-------------+
|Level     |Casting    |Range/Area |Components   |
|          |Time       |           |             |
+----------+-----------+-----------+-------------+
|1st       |1 Action   |30 ft      |V, S, M *    |
+----------+-----------+-----------+-------------+
|Duration  |School     |Attack/Save|Damage/Effect|
+----------+-----------+-----------+-------------+
|          |Enchantment|WIS Save   |Prone (...)  |
+----------+-----------+-----------+-------------+
```
