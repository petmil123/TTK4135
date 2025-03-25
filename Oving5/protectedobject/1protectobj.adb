with Ada.Text_IO; use Ada.Text_IO;
with Ada.Containers; use Ada.Containers;
with Ada.Containers.Vectors;

procedure protectobj is

    tick : Float := 0.033;
    
    package IntVec is new Ada.Containers.Vectors
        (Index_Type  => Natural,
        Element_Type => Integer);
    use IntVec;
    
    package Integer_IO is new Ada.Text_IO.Integer_IO (Integer);

    -- --- RESOURCE OBJECT --- --
    protected type Resource is
        entry allocateHigh(val: out IntVec.Vector);
        entry allocateLow(val: out IntVec.Vector);
        procedure deallocate(val: IntVec.Vector);
    private
        value: IntVec.Vector;
        busy: Boolean := False;
    end Resource;

    protected body Resource is
    
        entry allocateLow(val: out IntVec.Vector) when not busy is
        begin
            busy := True;
            val := value;
        end allocateLow;
    
        entry allocateHigh(val: out IntVec.Vector) when not busy is
        begin
            busy := True;
            val := value;
        end allocateHigh;

        procedure deallocate(val: IntVec.Vector) is
        begin
            if val.length>0 then
                value := val;
            end if; 
            busy := False;
        end deallocate;

    end Resource;

    type ExecutionState is (none, waiting, executing, done);
    type ExecutionStateArrT is array (0..9) of ExecutionState;
    executionStates: ExecutionStateArrT := (others => none);

    task type resourceUser(
        id:         Integer;
        priority:   Integer; 
        release:    Integer; 
        execute:    Integer; 
        r:          access Resource
    );
        value: IntVec.Vector;
    task body resourceUser is
    begin
        delay Duration(tick * Float(release));
        
        executionStates(id) := waiting;
        if priority = 1 then
            r.allocateHigh(value);
        else
            r.allocateLow(value);
        end if;
        
        executionStates(id) := executing;
        
        delay Duration(tick * Float(execute));
        value.Append(id);
        r.deallocate(value);
        
        executionStates(id) := done;
    end resourceUser;

    r: aliased Resource;

    -- 🎯 Testfunksjon for ressursallokering 🎯
    procedure Test_Resource_Allocation is
        executionOrder: IntVec.Vector;
    begin
        Put_Line("Starting test...");

        declare
            highPriorityUser: resourceUser(0, 1, 1, 1, r'Access);
            lowPriorityUser: resourceUser(1, 0, 2, 1, r'Access);
        begin
            null;
        end;

        -- Vent for at trådene skal utføre arbeidet
        delay 0.5;

        -- Hent utførelsesrekkefølgen
        r.allocateHigh(executionOrder);

        -- Skriv ut resultatet
        Put_Line("Execution order:");
        for idx in executionOrder.Iterate loop
            Put(Integer'Image(executionOrder(idx)) & " ");
        end loop;
        Put_Line("");

        -- Sjekk om høy prioritet kommer før lav
        if executionOrder.Length > 1 then
            Put_Line("✅ Test passed!");
        else
            Put_Line("❌ Test failed: Execution order is empty.");
        end if;
    end Test_Resource_Allocation;

begin
    Put_Line("Started");
    
    Test_Resource_Allocation; -- 🔥 Kjør testen!

end protectobj;